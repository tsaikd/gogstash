package inputfile

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/futil"
	"golang.org/x/sync/errgroup"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "file"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Path                 string `json:"path"`
	StartPos             string `json:"start_position,omitempty"` // one of ["beginning", "end"]
	SinceDBPath          string `json:"sincedb_path,omitempty"`
	SinceDBWriteInterval int    `json:"sincedb_write_interval,omitempty"`

	hostname            string
	SinceDBInfos        map[string]*SinceDBInfo `json:"-"`
	sinceDBLastInfosRaw []byte
	SinceDBLastSaveTime time.Time `json:"-"`
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		StartPos:             "end",
		SinceDBPath:          ".sincedb.json",
		SinceDBWriteInterval: 15,

		SinceDBInfos: map[string]*SinceDBInfo{},
	}
}

// errors
var (
	ErrorGlobFailed1 = errutil.NewFactory("glob(%q) failed")
)

// InitHandler initialize the input plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	logger := goglog.Logger

	if err = t.LoadSinceDBInfos(); err != nil {
		return
	}

	matches, err := filepath.Glob(t.Path)
	if err != nil {
		return ErrorGlobFailed1.New(err, t.Path)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return t.CheckSaveSinceDBInfosLoop(ctx)
	})

	for _, fpath := range matches {
		if fpath, err = evalSymlinks(ctx, fpath); err != nil {
			logger.Errorf("Get symlinks failed: %q\n%v", fpath, err)
			continue
		}

		var fi os.FileInfo
		if fi, err = os.Stat(fpath); err != nil {
			logger.Errorf("stat(%q) failed\n%s", t.Path, err)
			continue
		}

		if fi.IsDir() {
			logger.Infof("Skipping directory: %q", t.Path)
			continue
		}

		func(fpath string) {
			readEventChan := make(chan fsnotify.Event, 10)
			eg.Go(func() error {
				return t.fileReadLoop(ctx, readEventChan, fpath, msgChan)
			})
			eg.Go(func() error {
				return t.fileWatchLoop(ctx, readEventChan, fpath, fsnotify.Create|fsnotify.Write)
			})
		}(fpath)
	}

	return eg.Wait()
}

func (t *InputConfig) fileReadLoop(
	ctx context.Context,
	readEventChan chan fsnotify.Event,
	fpath string,
	msgChan chan<- logevent.LogEvent,
) (err error) {
	var (
		since     *SinceDBInfo
		fp        *os.File
		truncated bool
		ok        bool
		whence    int
		reader    *bufio.Reader
		line      string
		size      int

		buffer = &bytes.Buffer{}
		logger = goglog.Logger
	)

	if fpath, err = evalSymlinks(ctx, fpath); err != nil {
		logger.Errorf("Get symlinks failed: %q\n%v", fpath, err)
		return
	}

	if since, ok = t.SinceDBInfos[fpath]; !ok {
		t.SinceDBInfos[fpath] = &SinceDBInfo{}
		since = t.SinceDBInfos[fpath]
	}

	if since.Offset == 0 {
		if t.StartPos == "end" {
			whence = io.SeekEnd
		} else {
			whence = io.SeekStart
		}
	} else {
		whence = io.SeekStart
	}

	if fp, reader, err = openfile(fpath, since.Offset, whence); err != nil {
		return
	}
	defer fp.Close()

	if truncated, err = isFileTruncated(fp, since); err != nil {
		return
	}
	if truncated {
		logger.Warnf("File truncated, seeking to beginning: %q", fpath)
		since.Offset = 0
		if _, err = fp.Seek(since.Offset, io.SeekStart); err != nil {
			logger.Errorf("seek file failed: %q", fpath)
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if line, size, err = readline(ctx, reader, buffer); err != nil {
			if err == io.EOF {
				watchev := <-readEventChan
				logger.Debug("fileReadLoop recv:", watchev)
				if watchev.Op&fsnotify.Create == fsnotify.Create {
					logger.Warnf("File recreated, seeking to beginning: %q", fpath)
					fp.Close()
					since.Offset = 0
					if fp, reader, err = openfile(fpath, since.Offset, io.SeekStart); err != nil {
						return
					}
				}
				if truncated, err = isFileTruncated(fp, since); err != nil {
					return
				}
				if truncated {
					logger.Warnf("File truncated, seeking to beginning: %q", fpath)
					since.Offset = 0
					if _, err = fp.Seek(since.Offset, io.SeekStart); err != nil {
						logger.Errorf("seek file failed: %q", fpath)
						return
					}
					continue
				}
				logger.Debugf("watch %q %q %v", watchev.Name, fpath, watchev)
				continue
			} else {
				return
			}
		}

		_, err := t.Codec.Decode(ctx, []byte(line),
			map[string]any{
				"host":   t.hostname,
				"path":   fpath,
				"offset": since.Offset,
			},
			[]string{},
			msgChan)

		if err == nil {
			since.Offset += int64(size)
			if err := t.CheckSaveSinceDBInfos(); err != nil {
				return err
			}
		} else {
			logger.Errorf("Failed to decode %v using codec %v", line, t.Codec)
		}
	}
}

func (self *InputConfig) fileWatchLoop(ctx context.Context, readEventChan chan fsnotify.Event, fpath string, op fsnotify.Op) (err error) {
	var (
		event fsnotify.Event
	)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if event, err = waitWatchEvent(ctx, fpath, op); err != nil {
			return
		}
		readEventChan <- event
	}
}

func isFileTruncated(fp *os.File, since *SinceDBInfo) (truncated bool, err error) {
	var (
		fi os.FileInfo
	)
	if fi, err = fp.Stat(); err != nil {
		err = errutil.New("stat file failed: "+fp.Name(), err)
		return
	}
	if fi.Size() < since.Offset {
		truncated = true
	} else {
		truncated = false
	}
	return
}

func openfile(fpath string, offset int64, whence int) (fp *os.File, reader *bufio.Reader, err error) {
	if fp, err = os.Open(fpath); err != nil {
		err = errutil.New("open file failed: "+fpath, err)
		return
	}

	if _, err = fp.Seek(offset, whence); err != nil {
		err = errutil.New("seek file failed: " + fpath)
		return
	}

	reader = bufio.NewReaderSize(fp, 16*1024)
	return
}

func readline(ctx context.Context, reader *bufio.Reader, buffer *bytes.Buffer) (line string, size int, err error) {
	var (
		segment []byte
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if segment, err = reader.ReadBytes('\n'); err != nil {
			if err != io.EOF {
				err = errutil.New("read line failed", err)
			}
			return
		}

		if _, err = buffer.Write(segment); err != nil {
			err = errutil.New("write buffer failed", err)
			return
		}

		if isPartialLine(segment) {
			time.Sleep(1 * time.Second)
		} else {
			size = buffer.Len()
			line = buffer.String()
			buffer.Reset()
			line = strings.TrimRight(line, "\r\n")
			return
		}
	}
}

func isPartialLine(segment []byte) bool {
	if len(segment) < 1 {
		return true
	}
	if segment[len(segment)-1] != '\n' {
		return true
	}
	return false
}

var (
	mapWatcher = map[string]*fsnotify.Watcher{}
)

func waitWatchEvent(ctx context.Context, fpath string, op fsnotify.Op) (event fsnotify.Event, err error) {
	var (
		fdir    string
		watcher *fsnotify.Watcher
		ok      bool
	)

	if fpath, err = evalSymlinks(ctx, fpath); err != nil {
		err = errutil.New("Get symlinks failed: "+fpath, err)
		return
	}

	fdir = filepath.Dir(fpath)

	if watcher, ok = mapWatcher[fdir]; !ok {
		if watcher, err = fsnotify.NewWatcher(); err != nil {
			err = errutil.New("create new watcher failed: "+fdir, err)
			return
		}
		mapWatcher[fdir] = watcher
		if err = watcher.Add(fdir); err != nil {
			err = errutil.New("add new watch path failed: "+fdir, err)
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event = <-watcher.Events:
			if event.Name == fpath {
				if op > 0 {
					if event.Op&op > 0 {
						return
					}
				} else {
					return
				}
			}
		case err = <-watcher.Errors:
			err = errutil.New("watcher error", err)
			return
		}
	}
}

func evalSymlinks(ctx context.Context, path string) (string, error) {
	// https://github.com/tsaikd/gogstash/issues/30
	for retry := 5; retry > 0; retry-- {
		if futil.IsExist(path) {
			break
		}

		select {
		case <-ctx.Done():
			return path, nil
		case <-time.After(500 * time.Millisecond):
		}
	}

	if futil.IsNotExist(path) {
		return path, os.ErrNotExist
	}

	return filepath.EvalSymlinks(path)
}
