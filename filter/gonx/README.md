# gogstash gonx filter module

[Gonx](github.com/satyrius/gonx) is a log parser for gogstash. It is simple to use and can parse all kind of logfiles
that have a fixed text format. Gonx was originally written to parse nginx access log files. If you need more advanced parsing
then use the [grok](../grok) filter instead.

## Synopsis

```yaml
filter:
  - type: gonx
    # (optional) pattern to match, see below, default: nginx access log
    format: format-to-match
    # (optional) message field to parse, default: "message"
    source: "message"
```

## Example use

An example Apache log can look like this:

```text
127.0.0.1 - Scott [10/Dec/2019:13:55:36 -0700] "GET /server-status HTTP/1.1" 200 2326
```

To parse this with gonx a filter like this can be used:

```text
$ip $remotename $username [$time] "$request" $http_status $bytes
```

When parsed the variables will be fields in the event.

**NOTICE:**
1. If you using yaml config file, `\` should be written in `\\` in match patterns. For example: `"\\[%{HTTPDATE:nginx.access.time}\\]"`.
1. For JSON files you may also need to escape the quotes.
