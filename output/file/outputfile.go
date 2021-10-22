package outputfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	ossync "sync"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	fs "github.com/tsaikd/gogstash/output/file/filesystem"
)

const (
	// ModuleName is the name used in config file
	ModuleName               = "file"
	appendBehavior           = "append"
	overwriteBehavior        = "overwrite"
	createPerm               = os.O_CREATE | os.O_WRONLY
	appendPerm               = os.O_APPEND | os.O_WRONLY
	defaultWriteBehavior     = appendBehavior
	defaultDirMode           = "750"
	defaultFileMode          = "640"
	defaultFlushInterval     = 2
	defaultDiscardTime       = 300
	minDiscardTime           = 10
	defaultIdleCloseFileTime = 0
	defaultCodec             = "%{log}"
	defaultCreateIfDeleted   = true
)

// errors
var (
	ErrorNoPath               = errutil.NewFactory("no path defined for output file")
	ErrorInvalidWriteBehavior = errutil.NewFactory("invalid write_behavior defined: %s")
	ErrorInvalidFileMode      = errutil.NewFactory("invalid file_mode: %s")
	ErrorInvalidDirMode       = errutil.NewFactory("invalid dir_mode: %s")
	ErrorCreatingDir          = errutil.NewFactory("error creating directory: %s")
)

// osFS implements FileSystem using the local disk. this is the default implementation to use, except during when we mock
type osFS struct{}

func (osFS) Open(name string) (*os.File, error)    { return os.Open(name) }
func (osFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (osFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (osFS) OpenFile(name string, flag int, perm os.FileMode) (fs.File, error) {
	return os.OpenFile(name, flag, perm)
}

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	CreateIfDeleted bool   `json:"create_if_deleted"` // If the configured file is deleted, but an event is handled by the plugin, the plugin will recreate the file. Default ⇒ true
	DirMode         string `json:"dir_mode"`          // Dir access mode to use. Example: "dir_mode" => 0750
	FileMode        string `json:"file_mode"`         // File access mode to use. Example: "file_mode" => 0640
	FlushInterval   int    `json:"flush_interval"`    // Flush interval (in seconds) for flushing writes to log files. 0 will flush on every message.
	DiscardTime     int    `json:"discard_time"`      // Time (in seconds) for discarding messages before retrying to write to file
	IdleTimeout     int    `json:"idle_timeout"`      // Time (in seconds) without new messages before closing file and releasing resources. 0 will disable timeout.
	Path            string `json:"path"`              // The path to the file to write. Event fields can be used here, like /var/log/logstash/%{host}/%{application}
	Codec           string `json:"codec"`             // expression to write to file. E.g. "%{log}"
	WriteBehavior   string `json:"write_behavior"`    // If append, the file will be opened for appending and each new event will be written at the end of the file. If overwrite, the file will be truncated before writing and only the most recent event will appear in the file.
	fileMode        os.FileMode
	dirMode         os.FileMode
	fs              fs.FileSystem
	notifyDone      context.Context // notify all open files that we are shutting down

	writersMtx ossync.RWMutex         // mutex to protect writers
	writers    map[string]chan string // list of open files / active writers
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		CreateIfDeleted: defaultCreateIfDeleted,
		DirMode:         defaultDirMode,
		FileMode:        defaultFileMode,
		FlushInterval:   defaultFlushInterval,
		DiscardTime:     defaultDiscardTime,
		IdleTimeout:     defaultIdleCloseFileTime,
		WriteBehavior:   defaultWriteBehavior,
		Codec:           defaultCodec,
		writers:         make(map[string]chan string),
		fs:              osFS{},
	}
}

// parseAsIntOrOctal returns the parsed number, if it has a leading 0 it is parsed as octal, otherwise as decimal
func parseAsIntOrOctal(input string) (result int, err error) {
	base := 10
	if len(input) > 0 && input[0] == '0' {
		base = 8
	}
	i64, err := strconv.ParseInt(input, base, 32)
	result = int(i64)
	return
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}
	if conf.Path == "" {
		return nil, ErrorNoPath.New(nil)
	}
	if conf.WriteBehavior != appendBehavior && conf.WriteBehavior != overwriteBehavior {
		return nil, ErrorInvalidWriteBehavior.New(nil, conf.WriteBehavior)
	}
	var fMode, dMode int
	fMode, err = parseAsIntOrOctal(conf.FileMode)
	if err != nil || fMode < 0 {
		return nil, ErrorInvalidFileMode.New(err, conf.FileMode)
	}
	conf.fileMode = os.FileMode(fMode)
	dMode, err = parseAsIntOrOctal(conf.DirMode)
	if err != nil || dMode < 0 {
		return nil, ErrorInvalidDirMode.New(err, conf.DirMode)
	}
	conf.dirMode = os.FileMode(dMode)

	if conf.IdleTimeout < 0 {
		conf.IdleTimeout = 0
	}
	if conf.DiscardTime < minDiscardTime {
		conf.DiscardTime = minDiscardTime
	}
	if conf.FlushInterval < 0 {
		conf.FlushInterval = 0
	}

	var doneFunc func()
	conf.notifyDone, doneFunc = context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
			doneFunc()
		}
	}()

	return &conf, nil
}

func (t *OutputConfig) createFile(path string) (fs.File, error) {
	fileExists := t.exists(path)
	if !fileExists {
		dir := filepath.Dir(path)
		if dir != "." && !t.exists(dir) {
			if err := t.fs.MkdirAll(dir, t.dirMode); err != nil {
				return nil, ErrorCreatingDir.New(err, t.WriteBehavior)
			}
		}
	}
	var flags int
	switch t.WriteBehavior {
	case appendBehavior:
		if fileExists {
			flags = appendPerm
		} else {
			flags = createPerm
		}
	case overwriteBehavior:
		flags = createPerm
	default:
		return nil, ErrorInvalidWriteBehavior.New(nil, t.WriteBehavior)
	}
	return t.fs.OpenFile(path, flags, t.fileMode)
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	path := event.Format(t.Path)

	t.writersMtx.RLock()
	channel, alreadyWriting := t.writers[path]
	t.writersMtx.RUnlock()

	if !alreadyWriting {
		t.writersMtx.Lock()
		channel, alreadyWriting = t.writers[path]
		if alreadyWriting {
			t.writersMtx.Unlock()
		} else {
			channel = make(chan string, 100)
			t.writers[path] = channel
			t.writersMtx.Unlock()
			go func() {
				goglog.Logger.Debugf("Starting goroutine for file %s\n", path)

				// syncTick is used to sync file to operating system at regular intervals.
				// NewTicker can be recycled by the garbagecollector when Tick cannot.
				// A value of zero disables the ticker.
				var syncTick *time.Ticker
				if t.FlushInterval > 0 {
					syncTick = time.NewTicker(time.Duration(t.FlushInterval) * time.Second)
				} else {
					syncTick = time.NewTicker(time.Hour * 99999) // a ticker that will never tick
				}

				discardTimer := time.NewTimer(0) // used to handle discard timing
				discardTimer.Stop()

				const idleTickerInterval = 60                                // time in seconds between each check for idle timeout
				idleTick := time.NewTicker(idleTickerInterval * time.Second) // checks every idleTickerInterval

				isDiscarding := false // if true we are not writing to file but discarding incoming messages
				var discardCount uint // count number of messages discarded for reporting
				file, err := t.createFile(path)
				if err != nil {
					goglog.Logger.Errorf("problems opening %s %v, going into discard mode", path, err)
					isDiscarding = true
					discardCount = 1
					discardTimer.Reset(time.Duration(t.DiscardTime) * time.Second)
				}
				fileChanged := false  // true if file has changed (used with sync)
				var messageCount uint // count number of messages received for detection of idle timeout
				var idleCounter int   // count seconds of idle time
				for {
					select {
					case <-t.notifyDone.Done():
						goglog.Logger.Debugf("%s ctx.Done()", path)
						if !isDiscarding {
							err = closeFile(file)
							if err != nil {
								goglog.Logger.Errorf("%s: %v", path, err)
							}
						}
						return
					case <-idleTick.C:
						goglog.Logger.Debugf("%s idleTick", path)
						if messageCount > 0 {
							idleCounter = 0
						} else {
							idleCounter = idleCounter + idleTickerInterval
						}
						if idleCounter >= t.IdleTimeout {
							t.writersMtx.Lock()
							delete(t.writers, path)
							t.writersMtx.Unlock()
							goglog.Logger.Debugf("closing %s - idle timeout", path)
							if !isDiscarding {
								err = closeFile(file)
								if err != nil {
									goglog.Logger.Errorf("%s: %v", path, err)
								}
							}
							return
						}
						messageCount = 0
					case <-discardTimer.C:
						goglog.Logger.Debugf("%s discardTimer", path)
						if isDiscarding {
							goglog.Logger.Errorf("%s discarded %v messages, retrying", path, discardCount)
							isDiscarding = false
							discardCount = 0
						}
					case msg := <-channel:
						if isDiscarding {
							discardCount++
							continue
						}
						messageCount++
						msgToWrite := []byte(fmt.Sprintf("%s\n", msg))
						written, err := file.Write(msgToWrite)
						if err != nil {
							// error writing, close file, retry and go in discard mode if any error
							goglog.Logger.Errorf("%s: %v", path, err)
							closeFile(file)
							file, err = t.createFile(path)
							if err != nil {
								isDiscarding = true
								discardTimer.Reset(time.Duration(t.DiscardTime) * time.Second)
								goglog.Logger.Errorf("%s: %v - pausing file writing (discarding) for %v seconds", path, err, t.DiscardTime)
								discardCount = 1
								continue
							}
							written, err = file.Write(msgToWrite)
						}
						if written > 0 {
							fileChanged = true
						}
						if err != nil {
							goglog.Logger.Errorf("problems writing to %s %v. Written: %d", path, err, written)
							continue
						}
						goglog.Logger.Debugf("wrote %d bytes of %s to %s.", written, msg, path)
						if t.FlushInterval == 0 {
							sync(file)
						}
					case <-syncTick.C:
						if !fileChanged || isDiscarding {
							continue
						}
						goglog.Logger.Infof("sync'ing %s", path)
						sync(file)
						fileChanged = false
					}
				}
			}()
		}
	}

	log := event.Format(t.Codec)
	channel <- log
	return
}

func (t *OutputConfig) exists(filepath string) bool {
	if _, err := t.fs.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

// closeFile closes fs.File object by casting it to os.File. If it is not os.File an error is raised, if it is it returns the error code from Close()
func closeFile(file interface{}) (err error) {
	if myFile, isFile := file.(*os.File); isFile {
		err = myFile.Close()
	} else {
		msg := fmt.Sprintf("fileoutput closeFile got not *os.File object, got %v", reflect.TypeOf(file))
		err = errors.New(msg)
	}
	return
}

// sync sync file to disk
func sync(f fs.File) {
	err := f.Sync()
	if err != nil {
		goglog.Logger.Errorf("problems sync'ing %s: %v", f, err)
	}
}
