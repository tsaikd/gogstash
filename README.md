gogstash
========

Logstash like, written in golang

[![Build Status](https://travis-ci.org/tsaikd/gogstash.svg?branch=master)](https://travis-ci.org/tsaikd/gogstash)

* Download gogstash from github
	* [check latest version](https://github.com/tsaikd/gogstash/releases)
* Use docker image [tsaikd/gogstash](https://registry.hub.docker.com/u/tsaikd/gogstash/)

```
curl 'https://github.com/tsaikd/gogstash/releases/download/0.1.8/gogstash-Linux-x86_64' -SLo gogstash && chmod +x gogstash
```

* Configure for ubuntu-sys.json (example)
```
{
	"input": [
		{
			"type": "exec",
			"command": "sh",
			"interval": 60,
			"message_prefix": "%{@timestamp} [df] ",
			"args": ["-c", "df -B 1 / | sed 1d"]
		},
		{
			"type": "exec",
			"command": "sh",
			"interval": 60,
			"message_prefix": "%{@timestamp} [diskstat] ",
			"args": ["-c", "grep '0 [sv]da ' /proc/diskstats"]
		},
		{
			"type": "exec",
			"command": "sh",
			"interval": 60,
			"message_prefix": "%{@timestamp} [loadavg] ",
			"args": ["-c", "cat /proc/loadavg"]
		},
		{
			"type": "exec",
			"command": "sh",
			"interval": 60,
			"message_prefix": "%{@timestamp} [netdev] ",
			"args": ["-c", "grep '\\beth0:' /proc/net/dev"]
		},
		{
			"type": "exec",
			"command": "sh",
			"interval": 60,
			"message_prefix": "%{@timestamp} [meminfo]\n",
			"args": ["-c", "cat /proc/meminfo"]
		}
	],
	"output": [
		{
			"type": "report"
		},
		{
			"type": "redis",
			"key": "gogstash-ubuntu-sys-%{host}",
			"host": ["127.0.0.1:6379"]
		}
	]
}
```

* Configure for dockerstats.json (example)
```
{
	"input": [
		{
			"type": "dockerstats"
		}
	],
	"output": [
		{
			"type": "report"
		},
		{
			"type": "redis",
			"key": "gogstash-docker-%{host}",
			"host": ["127.0.0.1:6379"]
		}
	]
}
```

* Config format with YAML for dockerstats.json (example)
```
input:
  - type: dockerstats
output:
  - type: report
  - type: redis
    key: "gogstash-docker-%{host}"
    host:
      - "127.0.0.1:6379"
```

* Configure for nginx.yml with gonx filter (example)

```yml
chsize: 1000
worker: 2

input:
  - type: redis
    host: redis.server:6379
    key:  filebeat-nginx
    connections: 1

filter:
  - type: gonx
    format: '$clientip - $auth [$time_local] "$full_request" $response $bytes "$referer" "$agent"'
    source: message
  - type: gonx
    format: '$verb $request HTTP/$httpversion'
    source: full_request
  - type: date
    format: ["02/Jan/2006:15:04:05 -0700"]
    source: time_local
  - type: remove_field
    fields: ["full_request", "time_local"]
  - type: add_field
    key: host
    value: "%{beat.hostname}"
  - type: geoip2
    db_path: "GeoLite2-City.mmdb"
    ip_field: clientip
    key: req_geo
  - type: typeconv
    conv_type: int64
    fields: ["bytes", "response"]

output:
  - type: elastic
    url: "http://elastic.server:9200"
    index: "log-nginx-%{+@2006-01-02}"
    document_type: "%{type}"
```

* Configure for beats.yml with grok filter (example)

```yml
chsize: 1000
worker: 2
event:
  sort_map_keys: false
  remove_field: ['@metadata']

input:
  - type: beats
    port: 5044
    reuseport: true
    host: 0.0.0.0
    ssl:  false

filter:
  - type: grok
    match: ["%{COMMONAPACHELOG}"]
    source: "message"
    patterns_path: "/etc/gogstash/grok-patterns"
  - type: date
    format: ["02/Jan/2006:15:04:05 -0700"]
    source: time_local
  - type: remove_field
    fields: ["full_request", "time_local"]
  - type: add_field
    key: host
    value: "%{beat.hostname}"
  - type: geoip2
    db_path: "GeoLite2-City.mmdb"
    ip_field: clientip
    key: req_geo
  - type: typeconv
    conv_type: int64
    fields: ["bytes", "response"]

output:
  - type: elastic
    url: ["http://elastic1:9200","http://elastic2:9200","http://elastic3:9200"]
    index: "filebeat-6.4.2-%{+@2006.01.02}"
    document_type: "doc"
```

* Run gogstash for nginx example (command line)
```
GOMAXPROCS=4 ./gogstash --CONFIG nginx.json
```

* Run gogstash for dockerstats example (docker image)
```
docker run -it --rm \
	--name gogstash \
	--hostname gogstash \
	-e GOMAXPROCS=4 \
	-v "/var/run/docker.sock:/var/run/docker.sock" \
	-v "${PWD}/dockerstats.json:/gogstash/config.json:ro" \
	tsaikd/gogstash:0.1.8
```

## Supported inputs

See [input modules](input) for more information

* [beats](input/beats)
* [docker log](input/dockerlog)
* [docker stats](input/dockerstats)
* [exec](input/exec)
* [file](input/file)
* [http](input/http)
* [httplisten](input/httplisten)
* [nats](input/nats)
* [redis](input/redis)
* [socket](input/socket)

## Supported filters

All filters support the following commmon functionality/configuration:

```yaml
filter:
  - type: "whatever"

    # list of tags to add
    add_tag: ["addtag1", "addtag2"]
    
    # list of tags to remove
    remove_tag: ["removetag1", "removetag2"]
    
    # list of fields (key/value) to add
    add_field:
      - key: "field1"
        value: "value1"
      - key: "field2"
        value: "value2"
    # list of fields to remove
    remove_field: ["removefield1", "removefield2"]   
```

See [filter modules](filter) for more information

* [add field](filter/addfield)
* [cond](filter/cond)
* [date](filter/date)
* [geoip2](filter/geoip2)
* [gonx](filter/gonx)
* [grok](filter/grok)
* [json](filter/json)
* [mutate](filter/mutate)
* [rate limit](filter/ratelimit)
* [remove field](filter/removefield)
* [typeconv](filter/typeconv)
* [useragent](filter/useragent)

## Supported outputs

See [output modules](output) for more information

* [amqp](output/amqp)
* [cond](output/cond)
* [elastic](output/elastic)
* [email](output/email)
* [prometheus](output/prometheus)
* [redis](output/redis)
* [report](output/report)
* [socket](output/socket)
* [stdout](output/stdout)
