gogstash date filter module
=============================

## Synopsis

```yaml
filter:
  - type: date
    # (required) data format
    format: ["02/Jan/2006:15:04:05 -0700"]
    # (required) source field
    source: time_local
    # (optional) using joda time format, eg. "YYYY-MM-dd HH:mm:ss,SSS", default: false
    joda: false
    # (optional) target field. default: @timestamp
    target: mytimestamp
```
There are two special source formats:

* UNIX
* UNIXNANO

With these formats, the source is parsed and treated to be in unix time format with second or nanosecond precision.
