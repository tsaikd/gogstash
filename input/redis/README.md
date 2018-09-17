gogstash input redis
====================

## Synopsis

```yaml
input:
  # type Must be "redis"
  - type: "redis"

    # redis server host:port, default: "localhost:6379"
    host: "localhost:6379"

    # where to get data, default: "gogstash"
    key: "gogstash"

    # maximum number of socket connections, default: 10
    connections: 10

    # (optional) BLPOP blocking timeout, default: "600s"
    blocking_timeout: "600s"
```

## WARNING

redis client do not support golang context interface{} well, so interrupt signal from OS will not work
