package outputfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	fs "github.com/tsaikd/gogstash/output/file/filesystem"
)

const (
	// ModuleName is the name used in config file
	ModuleName             = "file"
	appendBehavior         = "append"
	overwriteBehavior      = "overwrite"
	createPerm             = os.O_CREATE | os.O_WRONLY
	appendPerm             = os.O_APPEND | os.O_WRONLY
	defaultWriteBehavior   = appendBehavior
	defaultDirMode         = "750"
	defaultFileMode        = "640"
	defaultFlushInterval   = 2
	defaultCodec           = "%{log}"
	defaultCreateIfDeleted = true
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
	Path            string `json:"path"`              // The path to the file to write. Event fields can be used here, like /var/log/logstash/%{host}/%{application}
	Codec           string `json:"codec"`             // expression to write to file. E.g. "%{log}"
	WriteBehavior   string `json:"write_behavior"`    // If append, the file will be opened for appending and each new event will be written at the end of the file. If overwrite, the file will be truncated before writing and only the most recent event will appear in the file.
	writers         map[string]chan string
	fileMode        os.FileMode
	dirMode         os.FileMode
	fs              fs.FileSystem
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
		WriteBehavior:   defaultWriteBehavior,
		Codec:           defaultCodec,
		writers:         make(map[string]chan string),
		fs:              osFS{},
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
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
	fMode, err = strconv.Atoi(conf.FileMode)
	if err != nil || fMode < 0 {
		return nil, ErrorInvalidFileMode.New(err, conf.FileMode)
	}
	conf.fileMode = os.FileMode(fMode)
	dMode, err = strconv.Atoi(conf.DirMode)
	if err != nil || dMode < 0 {
		return nil, ErrorInvalidDirMode.New(err, conf.DirMode)
	}
	conf.dirMode = os.FileMode(dMode)

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

	channel, alreadyWriting := t.writers[path]
	if !alreadyWriting {
		channel = make(chan string, 100)
		t.writers[path] = channel
		go func() {
			goglog.Logger.Debugf("Starting goroutine for file %s\n", path)
			var tick <-chan time.Time
			if t.FlushInterval > 0 {
				tick = time.Tick(time.Duration(t.FlushInterval) * time.Second)
			}
			file, err := t.createFile(path)
			if err != nil {
				goglog.Logger.Errorf("problems opening %s %v.\n", path, err)
				return
			}
			for {
				fileChanged := false
				select {
				case msg := <-channel:
					msgToWrite := []byte(fmt.Sprintf("%s\n", msg))
					written, err := file.Write(msgToWrite)
					if os.IsNotExist(err) && t.CreateIfDeleted {
						// re-create file if it was deleted and configured as such
						file, err = t.createFile(path)
						if err != nil {
							goglog.Logger.Errorf("problems re-creating file. Routine will not write anything else to file %s. %v.\n", path, err)
							return
						}
						written, err = file.Write(msgToWrite)
						if err != nil {
							goglog.Logger.Errorf("problems writting after re-creating file. Routine will not write anything else to file %s. %v.\n", path, err)
							return
						}
					}
					if written > 0 {
						fileChanged = true
					}
					if err != nil {
						goglog.Logger.Errorf("problems writing to %s %v. Written: %d\n", path, err, written)
						continue
					}
					goglog.Logger.Debugf("wrote %d bytes of %s to %s.\n", written, msg, path)
					if t.FlushInterval == 0 {
						sync(file)
					}
				case <-tick:
					if !fileChanged {
						continue
					}
					goglog.Logger.Infof("sync'ing %s\n", path)
					sync(file)
				}
			}
		}()
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

func sync(f fs.File) {
	err := f.Sync()
	if err != nil {
		goglog.Logger.Errorf("problems sync'ing %s: %v\n", f, err)
	}
}
