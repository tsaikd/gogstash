package inputfile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-fsnotify/fsnotify"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "file"
)

type InputConfig struct {
	config.InputConfig
	Path                 []string `json:"path"`
	StartPos             string   `json:"start_position,omitempty"` // one of ["beginning", "end"]
	SinceDBPath          string   `json:"sincedb_path,omitempty"`
	SinceDBWriteInterval int      `json:"sincedb_write_interval,omitempty"`

	hostname            string                  `json:"-"`
	SinceDBInfos        map[string]*SinceDBInfo `json:"-"`
	sinceDBLastInfosRaw []byte                  `json:"-"`
	SinceDBLastSaveTime time.Time               `json:"-"`
}

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

func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeInputConfig, err error) {
	conf := DefaultInputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if conf.hostname, err = os.Hostname(); err != nil {
		return
	}

	retconf = &conf
	return
}

func (t *InputConfig) Start() {
	t.Invoke(t.start)
}

func (t *InputConfig) start(logger *logrus.Logger, inchan config.InChan) (err error) {
	defer func() {
		if err != nil {
			logger.Errorln(err)
		}
	}()

	var (
		fi os.FileInfo
	)

	if err = t.LoadSinceDBInfos(); err != nil {
		return
	}
	matches := []string{}
	for _, v := range t.Path {
		var res []string
		if res, err = filepath.Glob(v); err != nil {
			return errutil.NewErrors(fmt.Errorf("glob(%q) failed", v), err)
		}
		matches = append(matches, res...)
	}

	go t.CheckSaveSinceDBInfosLoop()

	for _, fpath := range matches {
		if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
			logger.Errorf("Get symlinks failed: %q\n%v", fpath, err)
			continue
		}

		if fi, err = os.Stat(fpath); err != nil {
			logger.Errorf("stat(%q) failed\n%s", t.Path, err)
			continue
		}

		if fi.IsDir() {
			logger.Infof("Skipping directory: %q", t.Path)
			continue
		}

		readEventChan := make(chan fsnotify.Event, 10)
		go t.fileReadLoop(readEventChan, fpath, logger, inchan)
		go t.fileWatchLoop(readEventChan, fpath, logger, fsnotify.Create|fsnotify.Write)
	}

	return
}

func (t *InputConfig) fileReadLoop(
	readEventChan chan fsnotify.Event,
	fpath string,
	logger *logrus.Logger,
	inchan config.InChan,
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
	)

	if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
		logger.Errorf("Get symlinks failed: %q\n%v", fpath, err)
		return
	}

	if since, ok = t.SinceDBInfos[fpath]; !ok {
		t.SinceDBInfos[fpath] = &SinceDBInfo{}
		since = t.SinceDBInfos[fpath]
	}

	if since.Offset == 0 {
		if t.StartPos == "end" {
			whence = os.SEEK_END
		} else {
			whence = os.SEEK_SET
		}
	} else {
		whence = os.SEEK_SET
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
		if _, err = fp.Seek(since.Offset, os.SEEK_SET); err != nil {
			logger.Errorf("seek file failed: %q", fpath)
			return
		}
	}

	for {
		if line, size, err = readline(reader, buffer); err != nil {
			if err == io.EOF {
				watchev := <-readEventChan
				logger.Debug("fileReadLoop recv:", watchev)
				if watchev.Op&fsnotify.Create == fsnotify.Create {
					logger.Warnf("File recreated, seeking to beginning: %q", fpath)
					fp.Close()
					since.Offset = 0
					if fp, reader, err = openfile(fpath, since.Offset, os.SEEK_SET); err != nil {
						return
					}
				}
				if truncated, err = isFileTruncated(fp, since); err != nil {
					return
				}
				if truncated {
					logger.Warnf("File truncated, seeking to beginning: %q", fpath)
					since.Offset = 0
					if _, err = fp.Seek(since.Offset, os.SEEK_SET); err != nil {
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

		event := logevent.LogEvent{
			Timestamp: time.Now(),
			Message:   line,
			Extra: map[string]interface{}{
				"host":   t.hostname,
				"path":   fpath,
				"offset": since.Offset,
			},
		}

		since.Offset += int64(size)

		logger.Debugf("%q %v", event.Message, event)
		inchan <- event
		//self.SaveSinceDBInfos()
		t.CheckSaveSinceDBInfos()
	}
}

func (self *InputConfig) fileWatchLoop(readEventChan chan fsnotify.Event, fpath string, logger *logrus.Logger, op fsnotify.Op) (err error) {
	var (
		event fsnotify.Event
	)
	for {
		if event, err = waitWatchEvent(logger, fpath, op); err != nil {
			return
		}
		readEventChan <- event
	}
	return
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

func readline(reader *bufio.Reader, buffer *bytes.Buffer) (line string, size int, err error) {
	var (
		segment []byte
	)

	for {
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

	return
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

func waitWatchEvent(logger *logrus.Logger, fpath string, op fsnotify.Op) (event fsnotify.Event, err error) {
	var (
		watcher *fsnotify.Watcher
		ok      bool
	)

	if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
		err = errutil.New("Get symlinks failed: "+fpath, err)
		return
	}

	if watcher, ok = mapWatcher[fpath]; !ok {
		logger.Debugf("create new watcher for %q", fpath)
		if watcher, err = fsnotify.NewWatcher(); err != nil {
			err = errutil.New("create new watcher failed: "+fpath, err)
			return
		}
		mapWatcher[fpath] = watcher
		if err = watcher.Add(fpath); err != nil {
			err = errutil.New("add new watch path failed: "+fpath, err)
			return
		}
	}

	for {
		select {
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

	return
}
