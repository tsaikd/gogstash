package outputelastic

import (
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/olivere/elastic.v3"
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

	Sniff bool `json:"sniff"` // find all nodes of your cluster, https://github.com/olivere/elastic/wiki/Sniffing

	client *elastic.Client
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

	conf.client, err = elastic.NewClient(
		elastic.SetURL(conf.URL),
		elastic.SetSniff(conf.Sniff),
	)
	if err != nil {
		return
	}

	retconf = &conf
	return
}

func (t *OutputConfig) Event(event logevent.LogEvent) (err error) {
	index := event.Format(t.Index)
	doctype := event.Format(t.DocumentType)
	id := event.Format(t.DocumentID)

	_, err = t.client.Index().
		Index(index).
		Type(doctype).
		Id(id).
		BodyJson(event).
		Do()
	return
}
