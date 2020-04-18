gogstash input kafka
====================

## Synopsis

```yaml
input:
  # type Must be "kafka"
  - type: kafka

    # kafka version, (required)
    version: 0.10.2.0

    # kafka brokers host:port, (required)
    brokers:
      - 127.0.0.1:9092

    # topic for kafka client to listen, (required)
    topics:
      - testTopic

    # consumer group, (required)
    group: log_center

    # Kafka consumer consume initial offset from oldest
    offset_oldest: true

    # Consumer group partition assignment strategy (range, roundrobin)
    assignor: roundrobin

    # use SASL authentication (optional)
    security_protocol: SASL
    sasl_username: you-username
    sasl_password: you-password
```
