gogstash output redis
=====================

## Synopsis

```
{
	"output": [
		{
			"type": "redis",

			// (optional)
			"key": "gogstash",

			// (required)
			"host": [""],

			// (optional), default: "list"
			"data_type": "list",

			// (optional), in seconds, default: 5
			"timeout": 5

			// (optional), in seconds, default: 1
			"reconnect_interval": 1
		}
	]
}
```

## Details

* type
	* Must be **"redis"**
* key
	* The name of a Redis list or channel.
		Dynamic names are valid here, for example “gogstash-%{host}”.
* host
	* The hostname(s) and port(s) of your Redis server(s).
		Only one of redis server will be notify.
		When using redis master/slave, list all redis servers.

```
["127.0.0.1:6379", "10.20.30.40:6379"]
```

* timeout
	* Redis initial connection timeout in seconds.
* reconnect_interval
	* Interval for reconnecting to failed Redis connections.
