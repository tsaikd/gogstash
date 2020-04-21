gogstash input nats
====================

## Synopsis

```yaml
input:
  # type Must be "nats"
  - type: "nats"

    # nats server host:port, default: "nats://localhost:4222"
    host: "nats://localhost:4222"

    # creditials for nats, default: ""
    creds: ""

    # topics to subscribe, use , between topic
    topic: "test.*"

```
