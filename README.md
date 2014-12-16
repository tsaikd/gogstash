gogstash
==============

Logstash like, written in golang

* Download gogstash from github
```
curl 'https://github.com/tsaikd/gogstash/releases/download/0.0.2/gogstash-linux-amd64' -SLo gogstash && chmod +x gogstash
```

* Configure for nginx.json
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
