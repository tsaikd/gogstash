gogstash input http
===================

## Synopsis

```
{
	"input": [
		{
			"type": "http",

			// (optional), one of ["HEAD", "GET"], default: "GET"
			"method": "GET",

			// (required)
			"url": "",

			// (optional), in seconds, default: 60
			"interval": 60
		}
	]
}
```

## Details

* type
	* Must be **"http"**
* method
	* http request method
* url
	* http request url
* interval
	* How often (in seconds) to request a http endpoint.
