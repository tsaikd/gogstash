package outputelastic

import (
	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"
)

// ModuleName is the name used in config file
const ModuleName = "elastic"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URL          string `json:"url"`           // elastic API entrypoint
	Index        string `json:"index"`         // index name to log
	DocumentType string `json:"document_type"` // type name to log
	DocumentID   string `json:"document_id"`   // id to log, used if you want to control id format

	Sniff bool `json:"sniff"` // find all nodes of your cluster, https://github.com/olivere/elastic/wiki/Sniffing

	client *elastic.Client // elastic client instance
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the output plugin
func InitHandler(confraw *config.ConfigRaw, logger *logrus.Logger) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if conf.client, err = elastic.NewClient(
		elastic.SetURL(conf.URL),
		elastic.SetSniff(conf.Sniff),
	); err != nil {
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
		Do(context.TODO())
	return
}
