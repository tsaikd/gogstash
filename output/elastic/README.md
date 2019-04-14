gogstash output elastic
=======================

## Synopsis

```
output:
  - type: elastic
    # (required)
    # elastic API entrypoints
    url: ["http://127.0.0.1:9200"]

    # (required)
    # index name to log
    index: "gogstash-index-test"

    # (required)
    # type name to log
    document_type: "testtype"

    # (optional) default: ""
    # id to log, used if you want to control id format
    document_id: "%{fieldstring}"

    # (optional) default: false
    # find all nodes of your cluster, https://github.com/olivere/elastic/wiki/Sniffing
    sniff: true

    # (optional) default: 1000
    # BulkActions specifies when to flush based on the number of actions
    # currently added. Defaults to 1000 and can be set to -1 to be disabled.
    bulk_actions: 1000

    # (optional) default: 5242880 (5MB)
    # BulkSize specifies when to flush based on the size (in bytes) of the actions
    # currently added. Defaults to 5 MB and can be set to -1 to be disabled.
    bulk_size: 5242880

    # (optional) default: 30000000000 (30s)
    # BulkFlushInterval specifies when to flush at the end of the given interval.
    # Defaults to 30 seconds. If you want the bulk processor to
    # operate completely asynchronously, set both BulkActions and BulkSize to
    # -1 and set the FlushInterval to a meaningful interval.
    bulk_flush_interval: 30000000000

    # (optional) default: "10s"
    # ExponentialBackoffInitialTimeout used to set the first/minimal interval in elastic.ExponentialBackoff
    exponential_backoff_initial_timeout: "10s"

    # (optional) default: "5m"
    # ExponentialBackoffMaxTimeout used to set the maximum wait interval in elastic.ExponentialBackoff
    exponential_backoff_max_timeout: "5m"

    # (optional) default: "true"
    # SslCertValidation Option to validate the server's certificate. Disabling this severely compromises security. 
    # For more information on disabling certificate verification please read https://www.cs.utexas.edu/~shmat/shmat_ccs12.pdf     
    ssl_certificate_validation: false
```
