socket input
===================

Input event message should end with new line (`\n`) unless UDP packet mode is used.

## Synopsis

```
{
	"input": [
		{
			"type": "socket",

			// Socket type. Must be one of ["tcp", "udp", "unix", "unixpacket"].
			"socket": "tcp",

			// For TCP or UDP, address must have the form `host:port`.
			// For Unix networks, the address must be a file system path.
			"address": "localhost:9999",

			// (optional) SO_REUSEPORT applied or not, default: false
			"reuseport": false

			// (optional) sets UDP into packet mode, so each packet is processed individually.
			// In this mode the field "host_ip" is added with the source of the IP address in the format "host:port"
			"packetmode": true

			// (optional) set packet buffer size, for UDP this must be larger than the biggest packet you will receive.
			// Packets larger than this will be truncated down to the buffer size.
			"buffersize": 5000

			// (optional) codec that will process the incoming message. By default it will be processed as JSON,
			// if you want a different codec or the default (does nothing) you can configure this here.
			"codec": "default"
		}
	]
}
```

> Note: at the moment, UNIXGRAM socket are not supported.
