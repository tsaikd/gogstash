gogstash url_param filter module
=============================

This filter parses a url_params string in the log event message back to a map structure.
You can set a prefix for the Key name of the parsing result to avoid overwriting the original value due to duplicate names.

## Synopsis

```yaml
filter:
  - type: url_param
    # (optional) message field to parse, default: "request_url"
    source: request_url 
    # (optional) Which keys do you need to keep, default include all keys.
    include_keys: ["*"] 
    # (optional) Which key's value to do url_UnEscape, default is all key's value.
    url_decode: ["*"]
    # (optional) prefix for the result of Key name, cannot contain special characters.
    prefix: request_url_args_
    # (optional) remove the keys which value is empty string, default true.
    remove_empty_values: true
```

## Example for url_param 

Input string
```json
{
    "request_url": "http://example.com:80/path?start_time=2019-03-26%2000:00\u0026end_time=2019-03-27%2000:00\u0026isp=%E8%81%94%E9%80%9A\u0026area=%E5%8C%97%E4%BA%AC,%E9%87%8D%E5%BA%86,%E7%A6%8F%E5%BB%BA,%E7%94%98%E8%82%83\u0026audition_times=100\u0026metric=total\u0026combine=0\u0026api_addr=\u0026username=myName\u0026signature=ade6bd031b80ba5d4ed7e427c724796e"
}
```

Config:
```yaml
filter:
  - type: url_param
```

Produces the following output:
```json
{
    "request_url": "http://example.com:80/path?start_time=2019-03-26%2000:00\u0026end_time=2019-03-27%2000:00\u0026isp=%E8%81%94%E9%80%9A\u0026area=%E5%8C%97%E4%BA%AC,%E9%87%8D%E5%BA%86,%E7%A6%8F%E5%BB%BA,%E7%94%98%E8%82%83\u0026audition_times=100\u0026metric=total\u0026combine=0\u0026api_addr=\u0026username=myName\u0026signature=ade6bd031b80ba5d4ed7e427c724796e",
    "request_url_args_api_addr": "",
    "request_url_args_area": "北京,重庆,福建,甘肃",
    "request_url_args_audition_times": "100",
    "request_url_args_combine": "0",
    "request_url_args_end_time": "2019-03-27 00:00",
    "request_url_args_isp": "联通",
    "request_url_args_metric": "total",
    "request_url_args_signature": "ade6bd031b80ba5d4ed7e427c724796e",
    "request_url_args_start_time": "2019-03-26 00:00",
    "request_url_args_username": "myName"
}
```
