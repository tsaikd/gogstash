package outputemail

import (
	"crypto/tls"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/gomail.v2"
)

// ModuleName the module name of this plugin
const (
	ModuleName = "email"
)

// OutputConfig the default output config
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

// DefaultOutputConfig build the default output config
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

// InitHandler init the handler
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	retconf = &conf
	return
}

// Event the main log event
func (t *OutputConfig) Event(event logevent.LogEvent) (err error) {
	message := gomail.NewMessage()
	message.SetHeader("From", t.From)
	message.SetHeader("To", strings.Split(t.To, ";")...)
	if t.Cc != "" {
		message.SetHeader("Cc", strings.Split(t.Cc, ";")...)
	}
	message.SetHeader("Subject", t.Subject)

	if t.Attachments != nil {
		for _, v := range t.Attachments {
			message.Attach(v)
		}
	}

	message.SetBody("text/html", event.GetString("message"))
	dialer := gomail.NewDialer(t.Address, t.Port, t.UserName, t.Password)
	if t.UseTLS == true {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	err = dialer.DialAndSend(message)
	return
}
