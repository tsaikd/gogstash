package inputdocker

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var (
	containerLogMap = map[string]interface{}{}
	regNameTrim     = regexp.MustCompile(`^/`)
)

func (t *InputConfig) containerLogLoop(container interface{}, since *time.Time) (err error) {
	var (
		id   string
		name string
	)
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
		return errors.New("unsupported container type: " + reflect.TypeOf(container).String())
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
	stream := NewContainerLogStream(t.EventChan, id, eventExtra, since, nil)

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
