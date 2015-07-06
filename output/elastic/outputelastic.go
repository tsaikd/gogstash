package outputelastic

import (
	"net/url"
	"strings"

	"github.com/mattbaird/elastigo/lib"
	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "elastic"
)

type OutputConfig struct {
	config.OutputConfig
	URL          string `json:"url"`
	Index        string `json:"index"`
	DocumentType string `json:"document_type"`
	DocumentID   string `json:"document_id"`

	conn *elastigo.Conn
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	// elastic
	elasticurl, err := url.Parse(conf.URL)
	if err != nil {
		err = errutil.New("parse elastic url failed", err)
		return
	}

	conf.conn = elastigo.NewConn()
	conf.conn.Protocol = elasticurl.Scheme
	conf.conn.Domain = strings.Split(elasticurl.Host, ":")[0]
	conf.conn.Port = strings.Split(elasticurl.Host, ":")[1]
	if _, err = conf.conn.Health(); err != nil {
		err = errutil.New("test elastic connection failed", err)
		return
	}

	retconf = &conf
	return
}

func (t *OutputConfig) Event(event logevent.LogEvent) (err error) {
	index := event.Format(t.Index)
	doctype := event.Format(t.DocumentType)
	id := event.Format(t.DocumentID)
	if _, err = t.conn.Index(index, doctype, id, nil, event); err != nil {
		return
	}
	return
}
