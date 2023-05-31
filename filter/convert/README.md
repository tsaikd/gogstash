gogstash convert filter module
==============================

At least one of to_int, to_float
If merging does not succeed because destination field is not a string nor []string, gogstash_filter_convert_error will be added to the event

```yaml
filter:
  - type: convert
    # 2 item array. First item is field to convert. Second item is int64 to multiply the string converted by
    to_int: ["myfield", 1]
    # 3 item array. First item is field to convert. Second item is float64 to multiply the string converted by.
    to_float: ["myfield", 1]
```
