gogstash geoip2 filter module
=============================

You need to download geoip2 database manually and setup file path in config.

* http://dev.maxmind.com/geoip/geoip2/downloadable/
* http://dev.maxmind.com/geoip/geoip2/geolite2/

## Synopsis

```yaml
filter:
  - type: geoip2
    # (required) geoip2 database file path, default: GeoLite2-City.mmdb
    db_path: "geoip2/database/file/path/GeoLite2-City.mmdb"
    # (required) ip address field to parse
    ip_field: remote_addr
    # (optional) parsed geoip info should saved to field, default: geoip
    key: geoip
```
