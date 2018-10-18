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
```
