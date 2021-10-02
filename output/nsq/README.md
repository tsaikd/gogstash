# gogstash output NSQ

[NSQ](https://nsq.io) is a realtime distributed messaging platform.

## Synopsis

```yaml
output:
  - type: "nsq"

    # NSQd host:port to the daemon
    nsq: "localhost:4150"

    # topic is the topic to publish to
    topic: "mytopic"

    # (optional) The number of inflight messages to handle, default is 150
    max_inflight: 75
```
