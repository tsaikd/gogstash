package outputclickhouse

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sethgrid/pester"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "clickhouse"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	DSN               string        `json:"dsn"`                 // ClickHouse target URL including port and authentication. Eg. "http://default:changeme@mych:8123/"
	TargetTable       string        `json:"target_table"`        // Table where messages will be saved in, inluding database name. Eg. "my_database.my_table"
	BulkSize          int           `json:"bulk_size"`           // Size of bulk of messages to be sent at once
	BulkFlushInterval time.Duration `json:"bulk_flush_interval"` // Maximum time to wait between bulks
	Fields            []string      `json:"fields"`              // FIelds that will be sent to ClickHouse

	httpClient *pester.Client
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		DSN:               "http://localhost:8123",
		BulkSize:          1000,
		BulkFlushInterval: 5 * time.Second,
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	// Create http client.
	// @todo: allow user to set these parameters
	conf.httpClient = pester.New()
	conf.httpClient.Concurrency = 1
	conf.httpClient.MaxRetries = 0
	conf.httpClient.Backoff = pester.ExponentialBackoff
	conf.httpClient.KeepLog = true

	return &conf, nil
}

// Output event
// @todo: create bulk functionality
// @todo: add gzip functionality
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {

	newEventMap := make(map[string]interface{})

	for _, key := range t.Fields {
		fieldValue, exists := event.GetValue(key)

		if true == exists {
			switch fieldValue.(type) {
			case time.Time:
				timeValue, err := time.Parse(`2006-01-02 15:04:05.999999999 -0700 MST`, fmt.Sprintf("%s", fieldValue))
				if err != nil {
					log.Printf("Error parsing time field %s", err)
				}
				newEventMap[key] = timeValue.Format("2006-01-02 15:04:05")
			default:
				newEventMap[key] = fieldValue
			}
		}

	}

	newEventAsJSON, err := jsoniter.Marshal(newEventMap)

	if err != nil {
		return err
	}

	queryString := fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow", t.TargetTable)
	requestURL := fmt.Sprintf("%s?query=%s", t.DSN, url.QueryEscape(queryString))

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(string(newEventAsJSON)))
	resp, err := t.httpClient.Do(request)

	if err != nil {
		log.Println("Error executing request to Clickhouse: ", err)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal("cant read body", err)
		}

		log.Printf("response body %s", string(body))
	}

	return
}
