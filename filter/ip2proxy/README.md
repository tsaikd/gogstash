# gogstash ip2proxy filter module

You need to download ip2proxy database manually and setup file path in config.

* [Paying customers](https://www.ip2location.com/file-download)
* [Free databases](https://lite.ip2location.com)

The database will be reloaded when changed on the disk. The paid proxy database is updated daily and you should make
sure to download updates at regular intervals.

## Synopsis

```yaml
filter:
  - type: ip2proxy
    # (required) database file path
    db_path: "IP2PROXY-LITE-PX1.BIN"

    # (required) ip address field to parse
    ip_field: remote_addr

    # (optional) parsed ip2location info should be saved to field, default: ip2location
    key: ip2proxy

    # (optional) do not try to process private IP networks as they will fail, default: false
    skip_private: true

    # (optional) lets you specify your own definition for private IP addresses, both IPv4 and IPv6, default is private IP addresses
    private_net:
      - 10.0.0.0/8
      - 192.168.0.0/16

    # (optional) size of cache entries on IP addresses, so lookups don't go through the database, default is 100000
    cache_size: 100000

    # (optional) if true does not log lookup failures from the database, default is false
    quiet: true
```

Based on an input like this:
```json
{
  "ip": "1.1.1.1"
}
```

You should get an output like this with a proxy database:
```json
{
  "ip": "1.1.1.1",
  "ip2proxy": {
    "CountryLong": "United States of America",
    "CountryShort": "US",
    "ProxyType": "DCH",
    "isProxy": "2"
  }
}
```
