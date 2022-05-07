#gogstash add_field


## Synopsis

```yaml
filter:
  - type: add_field
    # (required) field to set
    key: field_to_set
    # (required) value
    value: value
    # (optional) set to true if existing field should be overwritten, default is false
    overwrite: true
```

This filter allows you to manually set a field to a value. There are a few points to have in mind:

1. The field "message" will always be overwritten if you set that as the key.
2. Other fields, if they exist, will not be overwritten unless you set ```overwrite: true```.
3. Value is formatted, meaning that
   1. a value of ```%{VALUE}``` will be replaced with the contents of VALUE if it exist. VALUE should then be another field defined from another filter or codec.
   2. ```%{+2006-01-02}``` will be replaced with the current time.
   3. ```%{+@2006-01-02}``` will be replaced with the event time.
