gogstash json filter module
=============================

This filter parses a JSON string in the log event message back to a map structure.
The JSON structure can either be appended to a new key or merged into the log event structure.

## Synopsis

```yaml
filter:
  - type: json
    # (optional) key of the parsed JSON map that contains the original log message
    message: msg
    # (optional) new key to append JSON map to
    appendkey: logmsg
    # (optional) key of the parsed JSON map that contains the log timestamp
    timestamp: time
    # (mandantory if timestamp is set) format string of the time
    timeformat: 2006-01-02T15:04:05.999999999Z
```

## Example for JSON append

Input string (from dockerlog input):
```
"{\"name\":\"Proxy\",\"hostname\":\"gogstash\",\"pid\":18,\"level\":30,\"module\":\"MQTT\",\"msg\":\"New close event for close.user.f30c2659cf56\",\"time\":\"2017-09-14T15:39:42.626Z\",\"v\":0}"
```

Config:
```yaml
filter:
  - type: json
    appendkey: logmsg
    timestamp: time
    timeformat: 2006-01-02T15:04:05.999999999Z
```

Produces the following output:
```json
{
	"@timestamp": "2017-09-14T15:39:42.626834631Z",
	"containerid": "ab5aa92191773848e2a1b28368b333e6afb5cc722ee4d243c8a754485aad8836",
	"containername": "proxy",
	"host": "localhost",
	"logmsg": {
		"hostname": "gogstash",
		"level": "30",
		"module": "MQTT",
		"msg": "New close event for close.user.f30c2659cf56",
		"name": "Proxy",
		"pid": 18,
		"time": "2017-09-14T15:39:42.626Z",
		"v": 0
    }
}
```

## Example for JSON merge

Input string (from dockerlog input):
```
"{\"name\":\"Proxy\",\"hostname\":\"gogstash\",\"pid\":18,\"level\":30,\"module\":\"MQTT\",\"msg\":\"New close event for close.user.f30c2659cf56\",\"time\":\"2017-09-14T15:39:42.626Z\",\"v\":0}"
```

Config:
```yaml
filter:
  - type: json
    message: msg
    timestamp: time
    timeformat: 2006-01-02T15:04:05.999999999Z
```

Produces the following output:
```json
{
	"@timestamp": "2017-09-14T15:39:42.626834631Z",
	"containerid": "ab5aa92191773848e2a1b28368b333e6afb5cc722ee4d243c8a754485aad8836",
	"containername": "proxy",
	"host": "localhost",
    "hostname": "gogstash",
    "level": "30",
    "module": "MQTT",
    "message": "New close event for close.user.f30c2659cf56",
    "name": "Proxy",
    "pid": 18,
    "time": "2017-09-14T15:39:42.626Z",
    "v": 0
}
```