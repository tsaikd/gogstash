gogstash input lorem
====================

## Synopsis

```yaml
input:
  # type Must be "lorem"
  - type: "lorem"

    # worker count to generate lorem, default: 1
    worker: 1

    # duration to generate lorem, set 0 to generate forever, default: 30s
    duration: "30s"

    # send empty messages without any lorem text, default: false
    empty: false
```
