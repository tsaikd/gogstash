# gogstash output http

## Synopsis

```yaml
output:
  - type: http
    # (required)
    # Endpoints to publish to (one in list is random chosen)
    urls: ["http://127.0.0.1/post"]

    # (optional)
    # HTTP status codes that indicates that message is delivered successfully, default list below
    http_status_codes: [200, 201, 202]

    # (optional)
    # HTTP status codes that indicates that message failed with an error where it is not possible to retry to send the
    # message, default list below
    http_error_codes: [501, 405, 404, 208, 505]

    # (optional)
    # How often to retry to send messages in case it was not delivered successfully, default is 30 seconds
    retry_interval: 30

    # (optional)
    # The maximum size of messages that can be queued (in case of delivery issues) before messages are dropped.
    # Default is one message, use -1 to not limit the queue.
    max_queue_size: 1

    # (optional)
    # true if you want to disable SSL/TLS validation of remote endpoint, default is false
    ignore_ssl: true
```
