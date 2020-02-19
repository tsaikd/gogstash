package inputdockerstats

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/input/dockerlog/dockertool"
)

var (
	containerMap = map[string]interface{}{}
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

	retry := 5
	statsChan := make(chan *docker.Stats, 100)

	for err == nil || retry > 0 {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case stats := <-statsChan:
					if time.Now().Add(-time.Duration(float64(t.StatInterval)-0.5) * time.Second).Before(*since) {
						continue
					}

					filterStatsByMode(stats, t.LogMode)

					extra := map[string]interface{}{
						"host":          t.hostname,
						"containerid":   id,
						"containername": name,
						"stats":         *stats,
					}

					*since = time.Now()

					t.Codec.Decode(ctx, "", extra, []string{}, msgChan)
				}
			}
		}()

		err = t.client.Stats(docker.StatsOptions{
			ID:      id,
			Stats:   statsChan,
			Stream:  true,
			Context: ctx,
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
		clear(&stats.PreCPUStats.CPUUsage.PercpuUsage)
		clear(&stats.PreCPUStats.CPUUsage.UsageInKernelmode)
		clear(&stats.PreCPUStats.CPUUsage.UsageInUsermode)
		clear(&stats.PreCPUStats.SystemCPUUsage)
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
