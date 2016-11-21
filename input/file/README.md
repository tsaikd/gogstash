gogstash input file
===================

## Synopsis

```
{
	"input": [
		{
			"type": "file",

			// (required)
			"path": [""],

			// (optional), one of ["beginning", "end"], default: "end"
			"start_position": "end",

			// (optional), default: ".sincedb.json"
			"sincedb_path": ".sincedb.json",

			// (optional), in seconds, default: 15
			"sincedb_write_interval": 15
		}
	]
}
```

## Details

* type
	* Must be **"file"**
* path
	* Path of file as input, seperated by line
* start_position
	* Choose where Logstash starts initially reading files:
		at the beginning or at the end.
		The default behavior treats files like live streams and thus starts at the end.
		If you have old data you want to import, set this to ‘beginning’
* sincedb_path
	* Where to write the sincedb database (keeps track of the current position of monitored log files).
* sincedb_write_interval
	* How often (in seconds) to write a since database with the current position of monitored log files.
