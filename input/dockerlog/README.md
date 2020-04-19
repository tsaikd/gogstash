gogstash input docker
=====================

## Synopsis

```
{
	"input": [
		{
			"type": "dockerlog",

			// (optional), docker endpoint, default: "unix:///var/run/docker.sock"
			"dockerurl": "unix:///var/run/docker.sock",

			// (optional), include docker name pattern, support regular expression of golang, default: []
			"include_patterns": [],

			// (optional), exclude docker name pattern, support regular expression of golang, default: ["gogstash"]
			"exclude_patterns": ["gogstash"],

			// (optional), sincedb storage path, default: "sincedb"
			"sincepath": "sincedb",

			// (optional), in seconds, docker connection retry interval, default: 10
			"connection_retry_interval": 10

			// (optional), path of TLS cert, default: ""
			"tls_cert": "/tmp/cert.pem"

			// (optional), path of TLS key, default: ""
			"tls_cert_key": "/tmp/key.pem"

			// (optional), path of TLS CA cert, default: ""
			"tls_ca_cert": "/tmp/ca.pem"
		}
	]
}
```
