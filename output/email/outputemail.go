package outputemail

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/gomail.v2"
)

// ModuleName is the name used in config file
const ModuleName = "email"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Address     string   `json:"address"`
	Attachments []string `json:"attachments"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Cc          string   `json:"cc"`
	Subject     string   `json:"subject"`
	Port        int      `json:"port"`
	UseTLS      bool     `json:"use_tls"`
	UserName    string   `json:"username"`
	Password    string   `json:"password"`
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Port:        25,
		UseTLS:      false,
		Cc:          "",
		UserName:    "",
		Password:    "",
		Attachments: nil,
	}
}

// InitHandler initialize the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	message := gomail.NewMessage()
	message.SetHeader("From", t.From)
	message.SetHeader("To", strings.Split(t.To, ";")...)
	if t.Cc != "" {
		message.SetHeader("Cc", strings.Split(t.Cc, ";")...)
	}
	message.SetHeader("Subject", t.Subject)

	if t.Attachments != nil && len(t.Attachments) > 0 {
		for _, v := range t.Attachments {
			message.Attach(v)
		}
	}

	message.SetBody("text/html", event.GetString("message"))
	dialer := gomail.NewDialer(t.Address, t.Port, t.UserName, t.Password)
	if t.UseTLS {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	err = dialer.DialAndSend(message)
	return
}
