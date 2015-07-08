package inputdockerlog

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	containerLogMap = map[string]interface{}{}
	regNameTrim     = regexp.MustCompile(`^/`)
)

func GetContainerInfo(container interface{}) (id string, name string, err error) {
	switch container.(type) {
	case docker.APIContainers:
		container := container.(docker.APIContainers)
		id = container.ID
		name = container.Names[0]
		name = regNameTrim.ReplaceAllString(name, "")
	case *docker.Container:
		container := container.(*docker.Container)
		id = container.ID
		name = container.Name
		name = regNameTrim.ReplaceAllString(name, "")
	default:
		err = errors.New("unsupported container type: " + reflect.TypeOf(container).String())
	}
	return
}

func (t *InputConfig) containerLogLoop(container interface{}, since *time.Time, evchan chan logevent.LogEvent, logger *logrus.Logger) (err error) {
	defer func() {
		if err != nil {
			logger.Errorln(err)
		}
		if recov := recover(); recov != nil {
			logger.Errorln(recov)
		}
	}()
	id, name, err := GetContainerInfo(container)
	if err != nil {
		return errutil.New("get container info failed", err)
	}
	if containerLogMap[id] != nil {
		return &ErrorContainerLogLoopRunning{id}
	}
	containerLogMap[id] = true
	defer delete(containerLogMap, id)

	eventExtra := map[string]interface{}{
		"host":          t.hostname,
		"containername": name,
	}

	retry := 5
	stream := NewContainerLogStream(evchan, id, eventExtra, since, nil)

	for err == nil || retry > 0 {
		err = t.client.Logs(docker.LogsOptions{
			Container:    id,
			OutputStream: &stream,
			ErrorStream:  &stream,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
			Timestamps:   true,
			Tail:         "",
			RawTerminal:  true,
		})
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			retry--
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}

	return
}
