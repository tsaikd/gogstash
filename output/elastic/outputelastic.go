package outputelastic

import (
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	elastic3 "gopkg.in/olivere/elastic.v3"
	elastic5 "gopkg.in/olivere/elastic.v5"
	"net/http"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"encoding/json"
	"github.com/hashicorp/go-version"
	"golang.org/x/net/context"
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
	ElasticVersion string `json:"es_version"` // 3, 5, or auto.

	Sniff bool `json:"sniff"` // find all nodes of your cluster, https://github.com/olivere/elastic/wiki/Sniffing

	client interface{} // we'll cast this to the proper client type when we're ready.
	clientVersion int // private var to hold client version to use after detection
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		ElasticVersion: "auto",
	}
}

func InitHandler(confraw *config.ConfigRaw, logger *logrus.Logger) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	switch (conf.ElasticVersion) {
	case "auto":
		conf.clientVersion = detectElasticVersion(&conf, logger)
	case "3":
		conf.clientVersion = 3
	case "5":
		conf.clientVersion = 5
	default:
		logger.Fatalf("Invalid config for es_version: expected one of [\"3\",\"5\",\"auto\"] but got: '%v'", conf.ElasticVersion)
	}


	if conf.clientVersion == 3 {
		conf.client, err = elastic3.NewClient(
			elastic3.SetURL(conf.URL),
			elastic3.SetSniff(conf.Sniff),
		)
	} else {
		conf.client, err = elastic5.NewClient(
			elastic5.SetURL(conf.URL),
			elastic5.SetSniff(conf.Sniff),
		)
	}

	if err != nil {
		return
	}

	retconf = &conf
	return
}

func detectElasticVersion(config *OutputConfig, logger *logrus.Logger) int {
	response, err := http.Get(config.URL)
	if err != nil {
		logger.Errorf("Unable to detect contact Elasticsearch to determine version. Error: %+v", err)
		return 0
	}
	defer response.Body.Close()
	buf, _ := ioutil.ReadAll(response.Body)

	var dest map[string]interface{}
	json.Unmarshal(buf, &dest)

	// yeah, i'd rather not make a struct just for this.
	ver, _ := version.NewVersion(dest["version"].(map[string]interface{})["number"].(string))

	v3constraint, _ := version.NewConstraint(">= 2.0.0, < 5.0.0")
	v5constraint, _ := version.NewConstraint(">= 5.0.0")

	if v3constraint.Check(ver) {
		logger.Debug("Detected Elasticsearch version 3")
		return 3
	} else if v5constraint.Check(ver) {
		logger.Debug("Detected Elasticsearch version 5")
		return 5
	} else {
		logger.Errorf("Unable to determine Elasticsearch version from version string: '%+v'", ver)
	}

	return 0

}

func (t *OutputConfig) Event(event logevent.LogEvent) (err error) {
	index := event.Format(t.Index)
	doctype := event.Format(t.DocumentType)
	id := event.Format(t.DocumentID)

	if t.clientVersion == 3 {
		_, err = t.client.(*elastic3.Client).Index().
			Index(index).
			Type(doctype).
			Id(id).
			BodyJson(event).
			Do()
	} else if t.clientVersion == 5 {
		_, err = t.client.(*elastic5.Client).Index().
			Index(index).
			Type(doctype).
			Id(id).
			BodyJson(event).
			Do(context.TODO())
	}
	return
}
