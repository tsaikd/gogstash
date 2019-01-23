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
    # (optional) parsed geoip info into flat format, default: false
    # `city_name`, `continent_code`, `country_code`, `country_name`,
    # `ip`, `latitude`, `longitude`, `postal_code`, `region_code`, `region_name` and `timezone`.
    flat_format: false
```
