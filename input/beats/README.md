gogstash input beats
====================

## Synopsis

```yaml
input:
  # type Must be "beats"
  - type: "beats"

    # (required) server port
    port: 5044

    # (optional) SO_REUSEPORT applied or not, default: false
    reuseport: false

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
