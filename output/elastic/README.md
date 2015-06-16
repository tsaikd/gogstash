gogstash output elastic
=======================

## Synopsis

```
{
	"output": [
		{
			"type": "elastic",

			// (required)
			"url": "http://127.0.0.1:9200",

			// (required)
			"index": "testindex",

			// (required)
			"document_type": "testtype",

			// (optional)
			"document_id": "%{fieldstring}"
		}
	]
}
```
