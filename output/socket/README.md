gogstash output socket
======================

## Synopsis

```
{
	"output": [
		{
			"type": "socket",

			// Socket type. Must be one of any of the types supported by net.Dial, i.e.: ["udp", "tcp", "unix", "unixpacket"].
			"socket": "tcp",

			// For TCP, address must have the form `host:port`.
			// For Unix networks, the address must be a file system path.
			"address": "localhost:9999"
		}
	]
}
```

## Details

* type
	* Must be **"socket"**
* socket
	* Socket type supported by [net.Dial](https://godoc.org/net#Dial)
* address
	* Address in a format supported by [net.Dial](https://godoc.org/net#Dial)
