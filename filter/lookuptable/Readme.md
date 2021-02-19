# Lookuptable

Translate key to value from file.

# Configuration

```yaml

filter:
  - type: lookuptable
    # Source field to match in lookup table, default: empty
    source: hostname
    # Target field where lookup value is set, default: empty
    target: service
    # File containing lookups, default: empty
    lookup_file: "hosts-lookuptable.txt"
    # (optional) Cache size, default: 1000
    cache_size: 1000
```

Lookup file contains one key-value pair per line.
Mapping must be in format <key>:<value>. All additional colons in key or value must be escaped with '\'.
If source-field matches lookup key, a target field is set to lookup value.

## Example lookup file
Map source hostname to service as per hosts-lookuptable.txt.
Use example yaml config block.

hosts-lookuptable.txt:
```text
192.168.100.10:webserver
192.168.100.20:db-server
192.168.100.20\:22:ssh-server
```

