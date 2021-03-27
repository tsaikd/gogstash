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

    # (optional) The number of events to return from Redis using EVAL, default: 125
    batch_count: 125

    # (optional) BLPOP blocking timeout, default: "600s"
    blocking_timeout: "600s"
    
    # (optional) Password for DB, default: null
    password: 
    
    # (option) Number of db to connect, default: 0
    db: 0
    
```

## WARNING

redis client do not support golang context interface{} well, so interrupt signal from OS will not work
