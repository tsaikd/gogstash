gogstash input dockerstats
==========================

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

			// (optional), exclude docker name pattern, support regular expression of golang, default: []
			"exclude_patterns": [],

			// (optional), in seconds, stat interval, default: 15
			"stat_interval": 15,

			// (optional), in seconds, docker connection retry interval, default: 10
			"connection_retry_interval": 10

			// (optional), zero hierarchical_memory_limit in docker stats, sometimes it will cause elastic exception 'JsonParseException Numeric value out of range of long', default: false
			"zero_hierarchical_memory_limit": false
		}
	]
}
```
