package modloader

import (
	"github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/filter/addfield"
	"github.com/tsaikd/gogstash/filter/cond"
	"github.com/tsaikd/gogstash/filter/date"
	"github.com/tsaikd/gogstash/filter/geoip2"
	"github.com/tsaikd/gogstash/filter/gonx"
	"github.com/tsaikd/gogstash/filter/grok"
	"github.com/tsaikd/gogstash/filter/json"
	"github.com/tsaikd/gogstash/filter/mutate"
	"github.com/tsaikd/gogstash/filter/ratelimit"
	"github.com/tsaikd/gogstash/filter/removefield"
	"github.com/tsaikd/gogstash/filter/typeconv"
	"github.com/tsaikd/gogstash/filter/urlparam"
	"github.com/tsaikd/gogstash/filter/useragent"
	"github.com/tsaikd/gogstash/input/beats"
	"github.com/tsaikd/gogstash/input/dockerlog"
	"github.com/tsaikd/gogstash/input/dockerstats"
	"github.com/tsaikd/gogstash/input/exec"
	"github.com/tsaikd/gogstash/input/file"
	"github.com/tsaikd/gogstash/input/http"
	"github.com/tsaikd/gogstash/input/httplisten"
	"github.com/tsaikd/gogstash/input/lorem"
	"github.com/tsaikd/gogstash/input/nats"
	"github.com/tsaikd/gogstash/input/redis"
	"github.com/tsaikd/gogstash/input/socket"
	"github.com/tsaikd/gogstash/output/amqp"
	"github.com/tsaikd/gogstash/output/cond"
	"github.com/tsaikd/gogstash/output/elastic"
	"github.com/tsaikd/gogstash/output/email"
	"github.com/tsaikd/gogstash/output/file"
	"github.com/tsaikd/gogstash/output/http"
	"github.com/tsaikd/gogstash/output/prometheus"
	"github.com/tsaikd/gogstash/output/redis"
	"github.com/tsaikd/gogstash/output/report"
	"github.com/tsaikd/gogstash/output/stdout"
)

func init() {
	config.RegistInputHandler(inputbeats.ModuleName, inputbeats.InitHandler)
	config.RegistInputHandler(inputdockerlog.ModuleName, inputdockerlog.InitHandler)
	config.RegistInputHandler(inputdockerstats.ModuleName, inputdockerstats.InitHandler)
	config.RegistInputHandler(inputexec.ModuleName, inputexec.InitHandler)
	config.RegistInputHandler(inputfile.ModuleName, inputfile.InitHandler)
	config.RegistInputHandler(inputhttp.ModuleName, inputhttp.InitHandler)
	config.RegistInputHandler(inputhttplisten.ModuleName, inputhttplisten.InitHandler)
	config.RegistInputHandler(inputlorem.ModuleName, inputlorem.InitHandler)
	config.RegistInputHandler(inputnats.ModuleName, inputnats.InitHandler)
	config.RegistInputHandler(inputredis.ModuleName, inputredis.InitHandler)
	config.RegistInputHandler(inputsocket.ModuleName, inputsocket.InitHandler)

	config.RegistFilterHandler(filteraddfield.ModuleName, filteraddfield.InitHandler)
	config.RegistFilterHandler(filtercond.ModuleName, filtercond.InitHandler)
	config.RegistFilterHandler(filterdate.ModuleName, filterdate.InitHandler)
	config.RegistFilterHandler(filtergeoip2.ModuleName, filtergeoip2.InitHandler)
	config.RegistFilterHandler(filtergonx.ModuleName, filtergonx.InitHandler)
	config.RegistFilterHandler(filtergrok.ModuleName, filtergrok.InitHandler)
	config.RegistFilterHandler(filterjson.ModuleName, filterjson.InitHandler)
	config.RegistFilterHandler(filtermutate.ModuleName, filtermutate.InitHandler)
	config.RegistFilterHandler(filterratelimit.ModuleName, filterratelimit.InitHandler)
	config.RegistFilterHandler(filterremovefield.ModuleName, filterremovefield.InitHandler)
	config.RegistFilterHandler(filtertypeconv.ModuleName, filtertypeconv.InitHandler)
	config.RegistFilterHandler(filteruseragent.ModuleName, filteruseragent.InitHandler)
	config.RegistFilterHandler(filterurlparam.ModuleName, filterurlparam.InitHandler)

	config.RegistOutputHandler(outputamqp.ModuleName, outputamqp.InitHandler)
	config.RegistOutputHandler(outputcond.ModuleName, outputcond.InitHandler)
	config.RegistOutputHandler(outputelastic.ModuleName, outputelastic.InitHandler)
	config.RegistOutputHandler(outputemail.ModuleName, outputemail.InitHandler)
	config.RegistOutputHandler(outputhttp.ModuleName, outputhttp.InitHandler)
	config.RegistOutputHandler(outputprometheus.ModuleName, outputprometheus.InitHandler)
	config.RegistOutputHandler(outputredis.ModuleName, outputredis.InitHandler)
	config.RegistOutputHandler(outputreport.ModuleName, outputreport.InitHandler)
	config.RegistOutputHandler(outputstdout.ModuleName, outputstdout.InitHandler)
	config.RegistOutputHandler(outputfile.ModuleName, outputfile.InitHandler)

	config.RegistCodecHandler(config.DefaultCodecName, config.DefaultCodecInitHandler)
	config.RegistCodecHandler(codecjson.ModuleName, codecjson.InitHandler)
}
