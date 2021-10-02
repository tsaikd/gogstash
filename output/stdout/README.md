gogstash output stdout
======================

## Synopsis

```json
{
	"output": [
		{
			"type": "stdout",
      "codec": "json"
		}
	]
}
```

## Details

Used for debug

* type
  * Must be **"stdout"**
* codec
  * Can be any supported codec that generates non-binary output. "default" prints the input message as it is and "json" prints the input as JSON.
