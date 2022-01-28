# gogstash discard filter module

This module allows you to define some conditions that will discard an event. This means that there will be no further processing of this event.
The syntax for the conditions are described in the [cond](../cond) filter.

You need to configure:

1. If the event should be discarded in case we have backpressure issues (the outputs cannot send the data out). The default is not to discard events, you need to enable it.
2. Your conditions. The event is discarded if *any* of the conditions evaluates to true.
3. If you want to negate the behaviour of the conditions; then the event will be discarded unless *any* of the conditions evaluates to true.

## Synopsis

```yaml
filter:
  - type: discard

    # (required unless you want to discard in case of backpressure)
    conditions:
      - "1==2"
      - "2==3"

    # (optional) if events should be discarded when we have a backpressure issue, default: false
    discard_if_backpressure: true

    # (optional) if we should negate the outcome of the conditions, default: false
    negate_conditions: true
```
