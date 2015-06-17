package inputdockerstats

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gogstash/config"
)

var (
	containerMap = map[string]interface{}{}
	regNameTrim  = regexp.MustCompile(`^/`)
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
	if containerMap[id] != nil {
		return &ErrorContainerLoopRunning{id}
	}
	containerMap[id] = true
	defer delete(containerMap, id)

	retry := 5

	for err == nil || retry > 0 {
		statsChan := make(chan *docker.Stats)

		go func() {
			for {
				select {
				case stats, ok := <-statsChan:
					if !ok {
						return
					}
					if time.Now().Add(-time.Duration(float64(t.StatInterval)-0.5) * time.Second).Before(*since) {
						continue
					}

					if t.ZeroHierarchicalMemoryLimit {
						stats.MemoryStats.Stats.HierarchicalMemoryLimit = 0
					}

					event := config.LogEvent{
						Timestamp: time.Now(),
						Extra: map[string]interface{}{
							"host":          t.hostname,
							"containerid":   id,
							"containername": name,
							"stats":         *stats,
						},
					}
					*since = time.Now()
					t.EventChan <- event
				}
			}
		}()

		err = t.client.Stats(docker.StatsOptions{
			ID:    id,
			Stats: statsChan,
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
