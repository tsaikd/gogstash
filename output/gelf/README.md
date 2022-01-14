# gogstash output GELF

[GELF](https://docs.graylog.org/docs/gelf) (Graylog Extended Log Format) is a Graylog log format

## Synopsis

```yaml
output:
  - type: "gelf"

    # List of Graylog Gelf Input. Format should be host:port
    hosts:
      - localhost:2022

    # (optional)
    # The maximum size of a chunk.
    # Default: 1420
    chunk_size: 1420

    # (optional)
    # The Level of compression.
    # Between 0 (None) to 9 (Best Compression).
    # Default: 1
    compression_level: 1420

    # (optional)
    # Type of compression.
    #  - Gzip: 0
    #  - Zlib: 1
    #  - None: 2
    # Default: 0 (Gzip)
    compression_type: 0

    # (optional)
    # How often to retry to send messages in case it was not delivered successfully
    # Default: 30 seconds
    retry_interval: 30

    # (optional)
    # The maximum size of messages that can be queued (in case of delivery issues) before messages are dropped.
    # Default is one message, use -1 to not limit the queue.
    max_queue_size: 1
```
