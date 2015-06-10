gogstash input docker
=====================

## Synopsis

```
{
	"input": [
		{
			"type": "docker",

			// (optional), host of docker endpoint, default: "unix:///var/run/docker.sock"
			"host": "unix:///var/run/docker.sock",

			// (optional), include docker name pattern, support regular expression of golang, default: []
			"include_patterns": [],

			// (optional), exclude docker name pattern, support regular expression of golang, default: ["gogstash"]
			"exclude_patterns": ["gogstash"],

			// (optional), sincedb storage path, default: "sincedb"
			"sincepath": "sincedb",

			// (optional), in seconds, docker connection retry interval, default: 10
			"connection_retry_interval": 10
		}
	]
}
```
