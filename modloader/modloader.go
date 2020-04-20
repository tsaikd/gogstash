package modloader

import (
	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	filteraddfield "github.com/tsaikd/gogstash/filter/addfield"
	filtercond "github.com/tsaikd/gogstash/filter/cond"
	filterdate "github.com/tsaikd/gogstash/filter/date"
	filtergeoip2 "github.com/tsaikd/gogstash/filter/geoip2"
	filtergonx "github.com/tsaikd/gogstash/filter/gonx"
	filtergrok "github.com/tsaikd/gogstash/filter/grok"
	filterjson "github.com/tsaikd/gogstash/filter/json"
	filtermutate "github.com/tsaikd/gogstash/filter/mutate"
	filterratelimit "github.com/tsaikd/gogstash/filter/ratelimit"
	filterremovefield "github.com/tsaikd/gogstash/filter/removefield"
	filtertypeconv "github.com/tsaikd/gogstash/filter/typeconv"
	filterurlparam "github.com/tsaikd/gogstash/filter/urlparam"
	filteruseragent "github.com/tsaikd/gogstash/filter/useragent"
	inputbeats "github.com/tsaikd/gogstash/input/beats"
	inputdockerlog "github.com/tsaikd/gogstash/input/dockerlog"
	inputdockerstats "github.com/tsaikd/gogstash/input/dockerstats"
	inputexec "github.com/tsaikd/gogstash/input/exec"
	inputfile "github.com/tsaikd/gogstash/input/file"
	inputhttp "github.com/tsaikd/gogstash/input/http"
	inputhttplisten "github.com/tsaikd/gogstash/input/httplisten"
	inputkafka "github.com/tsaikd/gogstash/input/kafka"
	inputlorem "github.com/tsaikd/gogstash/input/lorem"
	inputnats "github.com/tsaikd/gogstash/input/nats"
	inputredis "github.com/tsaikd/gogstash/input/redis"
	inputsocket "github.com/tsaikd/gogstash/input/socket"
	outputamqp "github.com/tsaikd/gogstash/output/amqp"
	outputcond "github.com/tsaikd/gogstash/output/cond"
	outputelastic "github.com/tsaikd/gogstash/output/elastic"
	outputelasticv5 "github.com/tsaikd/gogstash/output/elasticv5"
	outputemail "github.com/tsaikd/gogstash/output/email"
	outputfile "github.com/tsaikd/gogstash/output/file"
	outputhttp "github.com/tsaikd/gogstash/output/http"
	outputkafka "github.com/tsaikd/gogstash/output/kafka"
	outputprometheus "github.com/tsaikd/gogstash/output/prometheus"
	outputredis "github.com/tsaikd/gogstash/output/redis"
	outputreport "github.com/tsaikd/gogstash/output/report"
	outputsocket "github.com/tsaikd/gogstash/output/socket"
	outputstdout "github.com/tsaikd/gogstash/output/stdout"
)

func init() {
	config.RegistInputHandler(inputbeats.ModuleName, inputbeats.InitHandler)
	config.RegistInputHandler(inputdockerlog.ModuleName, inputdockerlog.InitHandler)
	config.RegistInputHandler(inputdockerstats.ModuleName, inputdockerstats.InitHandler)
	config.RegistInputHandler(inputexec.ModuleName, inputexec.InitHandler)
	config.RegistInputHandler(inputfile.ModuleName, inputfile.InitHandler)
	config.RegistInputHandler(inputhttp.ModuleName, inputhttp.InitHandler)
	config.RegistInputHandler(inputhttplisten.ModuleName, inputhttplisten.InitHandler)
	config.RegistInputHandler(inputkafka.ModuleName, inputkafka.InitHandler)
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
	config.RegistOutputHandler(outputelasticv5.ModuleName, outputelasticv5.InitHandler)
	config.RegistOutputHandler(outputemail.ModuleName, outputemail.InitHandler)
	config.RegistOutputHandler(outputhttp.ModuleName, outputhttp.InitHandler)
	config.RegistOutputHandler(outputprometheus.ModuleName, outputprometheus.InitHandler)
	config.RegistOutputHandler(outputredis.ModuleName, outputredis.InitHandler)
	config.RegistOutputHandler(outputreport.ModuleName, outputreport.InitHandler)
	config.RegistOutputHandler(outputsocket.ModuleName, outputsocket.InitHandler)
	config.RegistOutputHandler(outputstdout.ModuleName, outputstdout.InitHandler)
	config.RegistOutputHandler(outputfile.ModuleName, outputfile.InitHandler)
	config.RegistOutputHandler(outputkafka.ModuleName, outputkafka.InitHandler)

	config.RegistCodecHandler(config.DefaultCodecName, config.DefaultCodecInitHandler)
	config.RegistCodecHandler(codecjson.ModuleName, codecjson.InitHandler)
}
