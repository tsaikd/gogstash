# hash

This filter generates a hash out of your source fields and stores the hash in a target field.

```yaml
filter:
  - type: hash
    # (optional) source fields, default is [message]
    source: ["message"]
    # (optional) field to store the hash, default is hash
    target: hash
    # (optional) what kind of hash to create, default is sha1 - see list below
    kind: sha1
    # (optional) output format for hash, default is hex
    format: hex
```

## Supported hash formats (kind)

* adler32
* md5
* sha1
* sha256
* fnv32a
* fnv128a

## Supported output formats (format)

* base64
* binary
* int
* hex
