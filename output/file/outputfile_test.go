package outputfile

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	fs "github.com/tsaikd/gogstash/output/file/filesystem"
	mocks "github.com/tsaikd/gogstash/output/file/mocks"
)

func TestInvalidDefaultOutputConfig(t *testing.T) {
	assert := assert.New(t)

	// path not defined
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// write_behavior is different from 'append' and 'overwrite'
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: p
    write_behavior: write_behavior
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// invalid file_mode
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: p
    file_mode: -1
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// invalid file_mode
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: p
    file_mode: -1
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// invalid dir_mode
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: p
    dir_mode: -1
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// invalid flush_interval
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: p
    flush_interval: blah
	`)))
	assert.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.NotNil(err)

	// test default values
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
    output:
      - type: file
        path: p
        `)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	assert.Equal(defaultCreateIfDeleted, config.CreateIfDeleted)
	assert.Equal(defaultDirMode, config.DirMode)
	assert.Equal(defaultFileMode, config.FileMode)
	assert.Equal(defaultFlushInterval, config.FlushInterval)
	assert.Equal(defaultCodec, config.Codec)
	assert.Equal(defaultWriteBehavior, config.WriteBehavior)

}

func TestDefaultOutputConfigNewFile(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(640)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigNewFilePerm(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	fileMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(fileMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigCodec(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	fileMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
    codec: "test"
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(fileMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write([]byte("test")).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigCodecVar(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	fileMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
    codec: "%{log}"
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(fileMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	event.SetValue("log", "logvalue")
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write([]byte("logvalue")).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigCreateIfNeeded(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	fileMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(fileMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	event.SetValue("log", "logvalue")
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist).Times(2)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// first write, file will respond no error
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		return 10, nil
	})
	// second write, file will reply with 'ErrNotExist'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		return 0, os.ErrNotExist
	})
	// file will be opened twice. initial time and after error writing second time
	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	}).Times(2)
	err = config.Output(context.TODO(), event)
	assert.Nil(err)
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigCreateIfNeededDisabled(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	fileMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
    create_if_deleted: false
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(fileMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	event.SetValue("log", "logvalue")
	// filesystem will reply with 'file does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// first write, file will respond no error
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		return 10, nil
	})
	// second write, file will reply with 'ErrNotExist'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		return 0, os.ErrNotExist
	})
	// file will be opened twice. initial time and after error writing second time
	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigNewFileDir(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "dir/file"
	fileMode := 666
	dirMode := 777
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    file_mode: "` + strconv.Itoa(fileMode) + `"
    dir_mode: "` + strconv.Itoa(dirMode) + `"
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	fPerm := os.FileMode(fileMode)
	dPerm := os.FileMode(dirMode)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'dir does not exist'
	mockfs.EXPECT().Stat(path).Return(nil, os.ErrNotExist)
	mockfs.EXPECT().Stat("dir").Return(nil, os.ErrNotExist)
	mockfs.EXPECT().MkdirAll("dir", dPerm).Return(nil)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, fPerm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigExistingOverwrittenFile(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 1000
    write_behavior: overwrite
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(640)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'file exists'
	mockfs.EXPECT().Stat(path).Return("", nil)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		done <- true
		return 10, nil
	})

	mockfs.EXPECT().OpenFile(path, createPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}

func TestDefaultOutputConfigSync(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	path := "p"
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: file
    path: ` + path + `    
    flush_interval: 0
	`)))
	assert.Nil(err)
	c, err := InitHandler(context.TODO(), &conf.OutputRaw[0])
	assert.Nil(err)
	config := c.(*OutputConfig)
	perm := os.FileMode(640)
	// simulate dir does not exist. should be created with right permissions
	mockfs := mocks.NewMockFileSystem(ctrl)
	config.fs = mockfs
	event := logevent.LogEvent{}
	// filesystem will reply with 'file exists'
	mockfs.EXPECT().Stat(path).Return("", nil)
	mockfile := mocks.NewMockFile(ctrl)
	// channel to prevent test from finishing before gorouting writes to file
	done := make(chan bool)
	// file will reply with '10 bytes written'
	mockfile.EXPECT().Write(gomock.Any()).Return(10, nil)
	// file will be sync'd and we write to done channel
	mockfile.EXPECT().Sync().DoAndReturn(func() error {
		done <- true
		return nil
	})

	mockfs.EXPECT().OpenFile(path, appendPerm, perm).DoAndReturn(func(path string, flag int, perm os.FileMode) (fs.File, error) {
		return mockfile, nil
	})
	err = config.Output(context.TODO(), event)
	assert.Nil(err)

	// wait for done channel or 2 seconds delay, whatever happends first
	select {
	case <-done:
	case <-time.Tick(2 * time.Second):
	}

}
