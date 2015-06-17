package outputelastic

import (
	"net/url"
	"strings"

	"github.com/mattbaird/elastigo/lib"
	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "elastic"
)

type OutputConfig struct {
	config.CommonConfig
	URL          string `json:"url"`
	Index        string `json:"index"`
	DocumentType string `json:"document_type"`
	DocumentID   string `json:"document_id"`

	conn *elastigo.Conn
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
	}
}

func init() {
	config.RegistOutputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeOutputConfig, err error) {
		conf := DefaultOutputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
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
	})
}

func (t *OutputConfig) Event(event config.LogEvent) (err error) {
	index := event.Format(t.Index)
	doctype := event.Format(t.DocumentType)
	id := event.Format(t.DocumentID)
	if _, err = t.conn.Index(index, doctype, id, nil, event); err != nil {
		return
	}
	return
}
