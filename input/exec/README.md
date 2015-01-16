gogstash input exec
===================

## Synopsis

```
{
	"input": [
		{
			"type": "exec",

			// (required)
			"command": "",

			// (optional), default: []
			"args": [""],

			// (optional), in seconds, default: 60
			"interval": 60
		}
	]
}
```

## Details

* type
	* Must be **"exec"**
* command
	* Command to run. For example, “uptime”
* args
	* String array
	* Arguments of command
* interval
	* Interval to run the command. Value is in seconds.
