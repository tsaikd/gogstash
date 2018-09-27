gogstash input beats
====================

## Synopsis

```yaml
input:
  # type Must be "beats"
  - type: "beats"

    # (required) server port, SO_REUSEPORT applied
    port: 5044

    # (optional) server host, default: "0.0.0.0"
    host: "0.0.0.0"

    # (optional) Enable ssl transport, default: false
    ssl: false

    # SSL certificate file to use.
    #ssl_certificate:

    # SSL key file to use.
    #ssl_key:

    # SSL Verify, default: false
    #ssl_verify: false
```
