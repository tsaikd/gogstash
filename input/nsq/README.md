gogstash input NSQ
==================

[NSQ](https://nsq.io) is a realtime distributed messaging platform.

## Synopsis

```yaml
input:
  - type: "nsq"

    # NSQd port if connection directly to a daemon
    nsq: "localhost:4150"

    # lookupd if using the NSQ directory service
    lookupd: "server:4160"

    # topic is the topic to subscribe to
    topic: "mytopic"

    #channel is the channel you want messages from
    channel: "mychannel"

    # (optional) The number of inflight messages to handle, default is 75
    max_inflight: 75
```
