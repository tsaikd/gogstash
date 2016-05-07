package outputprometheus

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "prometheus"
)

type OutputConfig struct {
	config.OutputConfig
	Address string `json:"address,omitempty"`

	MsgCount prometheus.Counter `json:"-"`
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Address: ":8080",
		MsgCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "processed_messages_total",
			Help: "Number of processed messages",
		}),
	}
}

func InitHandler(confraw *config.ConfigRaw, logger *logrus.Logger) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	prometheus.MustRegister(conf.MsgCount)
	go conf.serveHTTP(logger)

	retconf = &conf
	return
}

func (o *OutputConfig) Event(event logevent.LogEvent) (err error) {
	o.MsgCount.Inc()
	return
}

func (o *OutputConfig) serveHTTP(logger *logrus.Logger) {
	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(o.Address, nil)
}
