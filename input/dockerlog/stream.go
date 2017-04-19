package inputdockerlog

import (
	"bytes"
	"io"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/gogstash/config/logevent"
)

func NewContainerLogStream(msgChan chan<- logevent.LogEvent, id string, eventExtra map[string]interface{}, since *time.Time, logger *logrus.Logger) ContainerLogStream {
	return ContainerLogStream{
		ID:         id,
		eventChan:  msgChan,
		eventExtra: eventExtra,
		logger:     logger,
		buffer:     bytes.NewBuffer(nil),

		since: since,
	}
}

type ContainerLogStream struct {
	io.Writer
	ID         string
	eventChan  chan<- logevent.LogEvent
	eventExtra map[string]interface{}
	logger     *logrus.Logger
	buffer     *bytes.Buffer
	since      *time.Time
}

func (t *ContainerLogStream) Write(p []byte) (n int, err error) {
	n, err = t.buffer.Write(p)
	if err != nil {
		t.logger.Fatal(err)
		return
	}

	idx := bytes.IndexByte(t.buffer.Bytes(), '\n')
	for idx > 0 {
		data := t.buffer.Next(idx)
		err = t.sendEvent(data)
		t.buffer.Next(1)
		if err != nil {
			t.logger.Fatal(err)
			return
		}
		idx = bytes.IndexByte(t.buffer.Bytes(), '\n')
	}
	return
}

var (
	reTime = regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+Z[0-9+-]*`)
)

func (t *ContainerLogStream) sendEvent(data []byte) (err error) {
	var (
		eventTime time.Time
	)

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   string(data),
		Extra:     t.eventExtra,
	}

	event.Extra["containerid"] = t.ID

	loc := reTime.FindIndex(data)
	if len(loc) > 0 && loc[0] < 10 {
		timestr := string(data[loc[0]:loc[1]])
		eventTime, err = time.Parse(time.RFC3339Nano, timestr)
		if err == nil {
			if eventTime.Before(*t.since) {
				return
			}
			event.Timestamp = eventTime
			data = data[loc[1]+1:]
		} else {
			t.logger.Println(err)
		}
	} else {
		t.logger.Printf("invalid event format %q\n", string(data))
	}

	event.Message = string(bytes.TrimSpace(data))

	if t.since.Before(event.Timestamp) {
		*t.since = event.Timestamp
	} else {
		return
	}

	if err != nil {
		event.AddTag("inputdocker_failed")
		err = nil
	}

	t.eventChan <- event

	return
}
