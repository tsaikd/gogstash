gogstash output statsd
=====================

## Synopsis

```
output:
  - type: cond
    condition: "([logmsg.message] ?? '') == 'access'"
    output:
      - type: statsd
        prefix: "Log.stage"
        increment:
          - "increment.%{logmsg.status}"
        decrement:
          - "decrement.%{logmsg.status}"
        count:
          - name: "count.%{logmsg.status}"
            value: "%{logmsg.count}"
        gauge:
          - name: "gauge.%{logmsg.status}"
            value: "%{logmsg.count}"
        timing:
          - name: "responce_time"
            value: "%{logmsg.time}"
          - name: "connect_time"
            value: "%{logmsg.connect_time}"
```

## Details

* type
	* Must be **"statsd"**
* host
	* The hostname(s) and port(s) of your StatsD server (by default 127.0.0.1:8125)
*	proto
  * Protocol for send StatsD metrics (udp, tcp) (by default udp).
* timeout
  * Timeout for connection/send to StatsD server (by default 5s).
* flush_interval
  * Flush interval (for bath sending to StatsD, by default 100ms).
* prefix
  * Prefix for sended metrics (without ending dot) (by default is empty)
*	increment
  * Increment template (like "name.status_code.%{logmsg.status} for lookup in logmsg[status]) (by default is disabled)
```
