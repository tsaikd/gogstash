gogstash grok filter module
=============================

## Synopsis

```yaml
filter:
  - type: grok
    # (optional) grok pattern, default: "%{COMMONAPACHELOG}"
    match: "%{COMMONAPACHELOG}"
    # (optional) message field to parse, default: "message"
    source: "message"
    # (optional) grok patterns file path, default: empty
    patterns_path: "path/to/file"
```
