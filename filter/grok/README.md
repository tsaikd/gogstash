gogstash grok filter module
=============================

## Synopsis

```yaml
filter:
  - type: grok
    # (optional) grok patterns, default: ["%{COMMONAPACHELOG}"]
    match: ["%{COMMONAPACHELOG}"]
    # (optional) message field to parse, default: "message"
    source: "message"
    # (optional) grok patterns dir path, default: empty
    patterns_path: "path/to/dir"
    # (optional) tags to add on grok success, default: empty
    add_tag: ["tag1", "tag2"]
    # (optional) tags to add on grok failure, default: ["gogstash_filter_grok_error"]
    tag_on_failure: ["errortag1", "errortag2"]
```
