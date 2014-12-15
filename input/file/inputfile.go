package inputfile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-fsnotify/fsnotify"

	"github.com/tsaikd/gogstash/config"
)

type InputConfig struct {
	config.CommonConfig
	Path                 string `json:"path"`
	StartPos             string `json:"start_position,omitempty"` // one of ["beginning", "end"]
	SinceDBPath          string `json:"sincedb_path,omitempty"`
	SinceDBWriteInterval int    `json:"sincedb_write_interval,omitempty"`

	EventChan           chan config.LogEvent    `json:"-"`
	SinceDBInfos        map[string]*SinceDBInfo `json:"-"`
	sinceDBLastInfosRaw []byte                  `json:"-"`
	SinceDBLastSaveTime time.Time               `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: "file",
		},
		StartPos:             "end",
		SinceDBPath:          ".sincedb.json",
		SinceDBWriteInterval: 15,

		SinceDBInfos: map[string]*SinceDBInfo{},
	}
}

func init() {
	config.RegistInputHandler("file", func(mapraw map[string]interface{}) (conf config.TypeInputConfig, err error) {
		var (
			raw []byte
		)
		if raw, err = json.Marshal(mapraw); err != nil {
			log.Error(err)
			return
		}
		defconf := DefaultInputConfig()
		conf = &defconf
		if err = json.Unmarshal(raw, &conf); err != nil {
			log.Error(err)
			return
		}
		return
	})
}

func (self *InputConfig) Type() string {
	return self.CommonConfig.Type
}

func (self *InputConfig) Event(eventChan chan config.LogEvent) (err error) {
	var (
		matches []string
		fi      os.FileInfo
	)

	if self.EventChan != nil {
		err = errors.New("Event chan already inited")
		log.Error(err)
		return
	}
	self.EventChan = eventChan

	if err = self.LoadSinceDBInfos(); err != nil {
		return
	}

	if matches, err = filepath.Glob(self.Path); err != nil {
		log.Errorf("glob(%q) failed\n%s", self.Path, err)
		return
	}

	go self.CheckSaveSinceDBInfosLoop()

	for _, fpath := range matches {
		if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
			log.Errorf("Get symlinks failed: %q\n%v", fpath, err)
			continue
		}

		if fi, err = os.Stat(fpath); err != nil {
			log.Errorf("stat(%q) failed\n%s", self.Path, err)
			continue
		}

		if fi.IsDir() {
			log.Infof("Skipping directory: %q", self.Path)
			continue
		}

		readEventChan := make(chan fsnotify.Event, 10)
		go self.fileReadLoop(readEventChan, fpath)
		go self.fileWatchLoop(readEventChan, fpath, fsnotify.Create|fsnotify.Write)
	}

	return
}

func (self *InputConfig) fileReadLoop(
	readEventChan chan fsnotify.Event,
	fpath string,
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
		hostname  string

		buffer = &bytes.Buffer{}
	)

	if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
		log.Errorf("Get symlinks failed: %q\n%v", fpath, err)
		return
	}

	if since, ok = self.SinceDBInfos[fpath]; !ok {
		self.SinceDBInfos[fpath] = &SinceDBInfo{}
		since = self.SinceDBInfos[fpath]
	}

	if since.Offset == 0 {
		if self.StartPos == "end" {
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
		log.Warnf("File truncated, seeking to beginning: %q", fpath)
		since.Offset = 0
		if _, err = fp.Seek(since.Offset, os.SEEK_SET); err != nil {
			log.Errorf("seek file failed: %q", fpath)
			return
		}
	}

	if hostname, err = os.Hostname(); err != nil {
		log.Errorf("Get hostname failed: %v", err)
	}

	for {
		if line, size, err = readline(reader, buffer); err != nil {
			if err == io.EOF {
				watchev := <-readEventChan
				log.Debug("fileReadLoop recv:", watchev)
				if watchev.Op&fsnotify.Create == fsnotify.Create {
					log.Warnf("File recreated, seeking to beginning: %q", fpath)
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
					log.Warnf("File truncated, seeking to beginning: %q", fpath)
					since.Offset = 0
					if _, err = fp.Seek(since.Offset, os.SEEK_SET); err != nil {
						log.Errorf("seek file failed: %q", fpath)
						return
					}
					continue
				}
				log.Debugf("watch %q %q %v", watchev.Name, fpath, watchev)
				continue
			} else {
				return
			}
		}

		event := config.LogEvent{
			Timestamp: time.Now(),
			Message:   line,
			Extra: map[string]interface{}{
				"host":   hostname,
				"path":   fpath,
				"offset": since.Offset,
			},
		}

		since.Offset += int64(size)

		go func(event config.LogEvent) {
			log.Debugf("%q %v", event.Message, event)
			self.EventChan <- event
			//self.SaveSinceDBInfos()
			self.CheckSaveSinceDBInfos()
			//chanSinceDBSave <- 1
		}(event)
	}

	return
}

func (self *InputConfig) fileWatchLoop(readEventChan chan fsnotify.Event, fpath string, op fsnotify.Op) (err error) {
	var (
		event fsnotify.Event
	)
	for {
		if event, err = waitWatchEvent(fpath, op); err != nil {
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
		log.Errorf("stat file failed: %q\n%s", fp.Name(), err)
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
	log.Debugf("openfile %q offset=%d whence=%d", fpath, offset, whence)
	if fp, err = os.Open(fpath); err != nil {
		log.Errorf("open file failed: %q\n%v", fpath, err)
		return
	}

	if _, err = fp.Seek(offset, whence); err != nil {
		log.Errorf("seek file failed: %q", fpath)
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
				log.Errorf("read line failed: %s", err)
			}
			return
		}

		if _, err = buffer.Write(segment); err != nil {
			log.Errorf("write buffer failed: %s", err)
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
	if len(segment) > 0 {
		if segment[len(segment)-1] == '\n' {
			return false
		}
	}
	return true
}

var (
	mapWatcher = map[string]*fsnotify.Watcher{}
)

func waitWatchEvent(fpath string, op fsnotify.Op) (event fsnotify.Event, err error) {
	var (
		fdir    string
		watcher *fsnotify.Watcher
		ok      bool
	)

	if fpath, err = filepath.EvalSymlinks(fpath); err != nil {
		log.Errorf("Get symlinks failed: %q\n%v", fpath, err)
		return
	}

	fdir = filepath.Dir(fpath)

	if watcher, ok = mapWatcher[fdir]; !ok {
		log.Debugf("create new watcher for %q", fdir)
		if watcher, err = fsnotify.NewWatcher(); err != nil {
			log.Errorf("create new watcher failed: %q\n%s", fdir, err)
			return
		}
		mapWatcher[fdir] = watcher
		if err = watcher.Add(fdir); err != nil {
			log.Errorf("add new watch path failed: %q\n%s", fdir, err)
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
			log.Errorf("watcher error: %s", err)
			return
		}
	}

	return
}
