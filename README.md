gogstash
========

Logstash like, written in golang

* Download gogstash from github
	* [check latest version](https://github.com/tsaikd/gogstash/releases)

```
curl 'https://github.com/tsaikd/gogstash/releases/download/0.0.4/gogstash-linux-amd64' -SLo gogstash && chmod +x gogstash
```

* Configure for nginx.json (example)
```
{
	"input": [
		{
			"type": "file",
			"path": "/var/log/nginx/access.log",
			"start_position": "beginning",
			"sincedb_path": ".sincedb.nginx.json"
		}
	],
	"filter": [],
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

* Run gogstash
```
GOMAXPROCS=4 ./gogstash --CONFIG nginx.json
```

## Supported inputs

* [exec](input/exec)
* [file](input/file)
* [http](input/http)

## Supported outputs

* [redis](output/redis)
* [report](output/report)
* [stdout](output/stdout)
