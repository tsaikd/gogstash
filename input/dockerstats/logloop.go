package inputdockerstats

import (
	"reflect"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/input/dockerlog"
)

var (
	containerMap = map[string]interface{}{}
)

func (t *InputConfig) containerLogLoop(container interface{}, since *time.Time, inchan config.InChan, logger *logrus.Logger) (err error) {
	defer func() {
		if err != nil {
			logger.Errorln(err)
		}
		if recov := recover(); recov != nil {
			logger.Errorln(recov)
		}
	}()
	id, name, err := inputdockerlog.GetContainerInfo(container)
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

					filterStatsByMode(stats, t.LogMode)

					event := logevent.LogEvent{
						Timestamp: time.Now(),
						Extra: map[string]interface{}{
							"host":          t.hostname,
							"containerid":   id,
							"containername": name,
							"stats":         *stats,
						},
					}
					*since = time.Now()
					inchan <- event
				}
			}
		}()

		err = t.client.Stats(docker.StatsOptions{
			ID:     id,
			Stats:  statsChan,
			Stream: true,
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

func filterStatsByMode(stats *docker.Stats, mode Mode) {
	switch mode {
	case ModeSimple:
		clearNetworkStats(&stats.Network)
		for name, network := range stats.Networks {
			clearNetworkStats(&network)
			stats.Networks[name] = network
		}
		clear(&stats.MemoryStats.Stats)
		clear(&stats.BlkioStats)
		clear(&stats.CPUStats.CPUUsage.PercpuUsage)
		clear(&stats.CPUStats.CPUUsage.UsageInKernelmode)
		clear(&stats.CPUStats.CPUUsage.UsageInUsermode)
		clear(&stats.CPUStats.SystemCPUUsage)
	}
}

func clearNetworkStats(network *docker.NetworkStats) {
	*network = docker.NetworkStats{
		RxBytes: network.RxBytes,
		TxBytes: network.TxBytes,
	}
}

func clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}
