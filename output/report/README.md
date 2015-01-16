gogstash output report
======================

## Synopsis

```
{
	"output": [
		{
			"type": "report",

			// (optional), in seconds, default: 5
			"interval": 5,

			// (optional), default: "[2/Jan/2006:15:04:05 -0700]"
			"time_format": "[2/Jan/2006:15:04:05 -0700]"
		}
	]
}
```

## Details

* type
	* Must be **"report"**
* interval
	* Interval for reporting event process count
* time_format
	* See [golang document](http://golang.org/pkg/time/) for more information.
