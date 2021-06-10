# KV

This filter converts a input string in a key=value format into fields.
The value kan be either quoted or without.
If the value is an integer it will be stored as an integer, otherwise a string.

```yaml
filter:
  - type: kv
    # (required) source field
    source: mysource
    # (optional) if target is not blank all fields will be placed under target
    target: mytarget
    # (optional) define what fields that should be string even it is an integer
    strings: ["x", "y"]
    # (optional) specify fields to remove after filtering
    remove_field: ["x", "y"]
```
