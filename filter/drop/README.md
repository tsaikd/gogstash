gogstash drop filter module
===========================

This filter stop the filter pipeline for this event and drop it

```yaml
filter:
  - type: cond
    condition: "..."
    filter:
      - type: drop
```
