package inputdockerstats

import (
	"log"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/input/docker"
)

var (
	containerMap = map[string]interface{}{}
)

func (t *InputConfig) containerLogLoop(container interface{}, since *time.Time) (err error) {
	defer func() {
		if err != nil {
			log.Println(err)
		}
	}()
	id, name, err := inputdocker.GetContainerInfo(container)
	if err != nil {
		return errutil.New("get container info failed", err)
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
