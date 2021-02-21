gogstash output email
=======================

## Synopsis

```
{
    "output": [
        {
            "type": "loki",

            // (required)
            "urls": ["http://192.168.1.11:3100/loki/api/v1/push"],

            // (optional)
            "auth": "loki_account:loki_passwd",
        }
    ]
}
```