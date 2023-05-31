gogstash input azure eventhub
=============================

## Synopsis

```yaml
input:
  # type Must be "azureeventhub"
  - type: azureeventhub

    # connection string to azure storage account used to store consumer group offsets (required)
    storage_connection_string: ""

    # container name to use to store consumer group offsets (required)
    storage_container: "testgogstash"
    
   # connection string to azure eventhub namespace (required)
    eventhub_namespace_connection_string: ""

    # eventhub name
    eventhub: evh-my-eventhub (required)

    # consumer group name
    group: gogstash
    
    # if not consumer group offset available, start from beginning (true) or last (false)
    offset_earliest: true

    # force constant attribut to every event
    extras:
      metadata_index_prefix: "prefix-"
```
