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

* Configure for nginx.json (example)
```
{
	"input": [
		{
			"type": "file",
			"path": ["/var/log/nginx/access.log"],
			"start_position": "beginning",
			"sincedb_path": ".sincedb.nginx.json"
		}
	],
	"output": [
		{
			"type": "report"
		},
		{
			"type": "redis",
			"key": "gogstash-nginx-%{host}",
			"host": ["127.0.0.1:6379"]
		}
	]
}
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

See [input module](input) for more information

* [docker log](input/dockerlog)
* [docker stats](input/dockerstats)
* [exec](input/exec)
* [file](input/file)
* [http](input/http)
* [socket](input/socket)

## Supported outputs

See [output module](output) for more information

* [amqp](output/amqp)
* [elastic](output/elastic)
* [redis](output/redis)
* [report](output/report)
* [stdout](output/stdout)
