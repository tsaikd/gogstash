package inputdockerlog

import (
	"context"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/input/dockerlog/dockertool"
)

func (t *InputConfig) containerLogLoop(ctx context.Context, container interface{}, since *time.Time, msgChan chan<- logevent.LogEvent) (err error) {
	id, name, err := dockertool.GetContainerInfo(container)
	if err != nil {
		return ErrorGetContainerInfoFailed.New(err)
	}
	if t.containerExist.Exist(id) {
		return ErrorContainerLoopRunning1.New(nil, id)
	}
	t.containerExist.Add(id)
	defer t.containerExist.Remove(id)

	eventExtra := map[string]interface{}{
		"host":          t.hostname,
		"containername": name,
	}

	retry := 5
	stream := NewContainerLogStream(msgChan, id, eventExtra, since, nil, t.Codec)

	for err == nil || retry > 0 {
		err = t.client.Logs(docker.LogsOptions{
			Context:      ctx,
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
