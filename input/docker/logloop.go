package inputdocker

import (
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gogstash/config"
)

var (
	containerLogMap = map[string]interface{}{}
)

func containerLogLoop(client *docker.Client, id string, eventChan chan config.LogEvent, eventExtra map[string]interface{}, since *time.Time) (err error) {
	if containerLogMap[id] != nil {
		return &ErrorContainerLogLoopRunning{id}
	}
	containerLogMap[id] = true
	defer delete(containerLogMap, id)

	retry := 5
	stream := NewContainerLogStream(eventChan, id, eventExtra, since, nil)

	for err == nil || retry > 0 {
		err = client.Logs(docker.LogsOptions{
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
