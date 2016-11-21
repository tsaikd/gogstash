package modloader

import (
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/filter/addfield"
	"github.com/tsaikd/gogstash/input/dockerlog"
	"github.com/tsaikd/gogstash/input/dockerstats"
	"github.com/tsaikd/gogstash/input/exec"
	"github.com/tsaikd/gogstash/input/file"
	"github.com/tsaikd/gogstash/input/http"
	"github.com/tsaikd/gogstash/input/socket"
	"github.com/tsaikd/gogstash/output/amqp"
	"github.com/tsaikd/gogstash/output/elastic"
	"github.com/tsaikd/gogstash/output/email"
	"github.com/tsaikd/gogstash/output/prometheus"
	"github.com/tsaikd/gogstash/output/redis"
	"github.com/tsaikd/gogstash/output/report"
	"github.com/tsaikd/gogstash/output/stdout"
)

func init() {
	config.RegistInputHandler(inputexec.ModuleName, inputexec.InitHandler)
	config.RegistInputHandler(inputdockerlog.ModuleName, inputdockerlog.InitHandler)
	config.RegistInputHandler(inputdockerstats.ModuleName, inputdockerstats.InitHandler)
	config.RegistInputHandler(inputfile.ModuleName, inputfile.InitHandler)
	config.RegistInputHandler(inputhttp.ModuleName, inputhttp.InitHandler)
	config.RegistInputHandler(inputsocket.ModuleName, inputsocket.InitHandler)

	config.RegistFilterHandler(filteraddfield.ModuleName, filteraddfield.InitHandler)

	config.RegistOutputHandler(outputstdout.ModuleName, outputstdout.InitHandler)
	config.RegistOutputHandler(outputelastic.ModuleName, outputelastic.InitHandler)
	config.RegistOutputHandler(outputredis.ModuleName, outputredis.InitHandler)
	config.RegistOutputHandler(outputreport.ModuleName, outputreport.InitHandler)
	config.RegistOutputHandler(outputamqp.ModuleName, outputamqp.InitHandler)
	config.RegistOutputHandler(outputprometheus.ModuleName, outputprometheus.InitHandler)
	config.RegistOutputHandler(outputemail.ModuleName, outputemail.InitHandler)
}
