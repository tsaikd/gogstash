gogstash output file
======================

## Synopsis

```
{
    "output": [{
        "type" : "file"
        "create_if_deleted" : true
        "dir_mode" : "750"
        "file_mode" : "640"
        "flush_interval" : 2
        "path" : "myfile.log"
        "codec" : "%{log}"
        "write_behavior" : "append"
    }]
}
```

## Details

* type
    * Must be **"file"**
* create_if_deleted
    * Optional boolean value. Default is true. Whether file will be re-created if it's deleted and output plugin receives new events.
* dir_mode
    * Optional string value. efault is "750". Permissions given to directory if it is created by this output plugin
* file_mode
    * Optional string value. Default is "640". Permissions given to file if it is created by this output plugin
* flush_interval
    * Optional number value. Default is 2. File sync rate, in seconds. File will only be sync'd if new events were received since last sync.
* path
    * Mandatory string value. Path of the file to write to. Accepts event variables, e.g. "file%{var}.log"
* codec
    * Optional string value. Default is "%{log}". Expression to write to file.
* write_behavior
    * Optional value, must be either "append" or "overwrite". Default is "append". Whether to append to existing files or overwrite them.
