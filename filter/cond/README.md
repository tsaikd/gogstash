gogstash cond filter module
=============================

Condition syntax depends `govaluate`:

* https://github.com/Knetic/govaluate

Only boolean `true` for the result value is considered to be eligible.

Built-in functions:

* `empty()` Checks if the argument is `nil`
* `strlen()` Returns the string argument's length
* `map()` Maps slice to `[]interface{}`
* `rand()` Returns a pseudo-random number in `[0.0, 1.0)` as a float64

## Synopsis

```yaml
filter:
  - type: cond
    # (required) condition need to be satisfied
    condition: "level == 'ERROR'"
    # (required) filter config
    filter:
      - type: add_field
        key: foo
        value: bar
    # (optional) filter config when condition was not met
    else_filter:
      - type: add_field
        key: foo
        value: bar2
```

```yaml
filter:
  - type: cond
    # (required) condition need to be satisfied
    condition: "!('gogstash_filter_grok_error' IN map([tags]))"
    # (required) filter config
    filter:
      - type: remove_field
        remove_message: true
```
