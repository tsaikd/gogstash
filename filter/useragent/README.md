gogstash useragent filter module
=============================

This filter parses a useragent string in the log event message to a structured form.

## Synopsis

```yaml
filter:
  - type: useragent
    # message field to parse
    source: headers.user_agent
    # target field
    target: user_agent
    # (optional) YAML file of user agent regexes to match. Default: regexes from uaparser
    # You can find the latest version of this here:
    # <https://github.com/ua-parser/uap-core/blob/master/regexes.yaml>
    regexes: "./regex.yaml"
    # (optional) Number of entries to cache, to avoid re-parsing the same user agents repeatedly. Default: 100000
    cache_size: 1000
```

## Example for useragent

Input string
```json
{
    "headers": {
        "user_agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
    }
}
```

Config:
```yaml
filter:
  - type: useragent
    source: headers.user_agent
    target: user_agent
```

Produces the following output:
```json
{
    "headers": {
        "user_agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
    },
    "user_agent": {
        "device": "Other",
        "major": "59",
        "minor": "0",
        "name": "Chrome",
        "os": "Windows",
        "os_major": "7",
        "os_name": "Windows",
        "patch": "3071"
    }
}
```
