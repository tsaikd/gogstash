gogstash cond elastic
=======================

Condition syntax is same as [`filter/cond`](../../filter/cond).

## Synopsis

```
{
	"output": [
		{
			"type": "cond",

			// (required)
			"condition": "level == 'ERROR'",

			// (required)
			"output": [

				// output config
				{ "type": "stdout" }

			]
		}
	]
}
```
