gogstash input dockerstats
==========================

## Synopsis

```
{
	"input": [
		{
			"type": "docker",

			// (optional), docker endpoint, default: "unix:///var/run/docker.sock"
			"dockerurl": "unix:///var/run/docker.sock",

			// (optional), include docker name pattern, support regular expression of golang, default: []
			"include_patterns": [],

			// (optional), exclude docker name pattern, support regular expression of golang, default: []
			"exclude_patterns": [],

			// (optional), in seconds, stat interval, default: 15
			"stat_interval": 15,

			// (optional), in seconds, docker connection retry interval, default: 10
			"connection_retry_interval": 10

			// (optional), filter the output by mode, available value: "full" | "simple", "simple" mode will remove some messages to make the log smaller, default: "full"
			"log_mode": "full"
		}
	]
}
```
