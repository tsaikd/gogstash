gogstash input httplisten
=========================

## Synopsis

```
{
	"input": [
		{
			"type": "httplisten",

			// (optional), hostIP:port, default: "localhost:8080"
			"address": "0.0.0.0:8080",

			// (optional), path to accept POST request, default: "/"
			"path": "/mypath/",

			// (optional), Server Certicate File including path, default: ""
			"cert": "/home/user/server.crt"

			// (optional), Server Key File including path. default: "", when both Certicate and Key files provided, HTTP server will start in TLS mode.
			"key": "/home/user/server.key"
		}
	]
}
```

## Details

* type
	* Must be **"httplisten"**
* address
	* http listening IP address and port
* path
	* http request path to POST request
* cert
	* Used for https (TLS) mode. Server Certicate File including path
* key
	* Server Key
