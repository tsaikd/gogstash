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
    # true if you want to disable SSL/TLS validation of remote endpoint, default is false
    ignore_ssl: true
```
