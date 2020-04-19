gogstash output kafka
====================

## Synopsis

```yaml
output:
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

    # use SASL authentication (optional)
    security_protocol: SASL
    sasl_username: you-username
    sasl_password: you-password
```
