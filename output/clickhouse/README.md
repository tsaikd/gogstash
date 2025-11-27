gogstash output clickhouse
=======================

## Synopsis

This plugin sends gogstash events to a ClickHouse table using the HTTP interface
and the `JSONEachRow` format.

Each event is converted to a single JSON object (one per line) and inserted with
a statement like:

```sql
INSERT INTO logs.ids FORMAT JSONEachRow
{"ts":"2025-11-21T12:34:56Z","host":"fw01","message":"...","level":"info"}
{"ts":"2025-11-21T12:34:57Z","host":"fw01","message":"...","level":"warn"}
...
```


## Configuration:

```
output:
  - type: clickhouse

    # List of ClickHouse HTTP endpoints. (required)
    # The plugin picks one randomly per flush, distributing load automatically.
    urls: ["http://clickhouse1:8124"]

    # Full table name: "database.table". (required)
    table: "db.table"

    # HTTP Basic Authentication. Default: "" (optional)
    auth: "gogstash:strong_password"

    # Flush when this many events are buffered. Default: 1000 (optional)
    batch_size: 2000

    # Max wait time before flushing. Default: "2s" (optional)
    flush_interval: "1s"

    # Name of an event field to map into ts column. Default: "" (optional - recommended)
    # Use field "@timestamp" from event. Extra as the "ts" column. 
    # If ts_field is not specified:
    # - the JSON rows do not include "ts"
    # - your ClickHouse table should have ts DateTime DEFAULT now()
    ts_field: "@timestamp"

    # If true, non-2xx ClickHouse status codes become plugin errors. Default: false (optional)
    # false → log error only (pipeline continues)
    # true → return non-nil error (pipeline may halt or retry)
    fail_on_error: true 

    # If true, the HTTP client will accept self-signed / invalid certificates. Default: false (optional)
    ssl_insecure_skip_verify: false

```
