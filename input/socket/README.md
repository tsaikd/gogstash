socket input
===================

## Synopsis

```
{
	"input": [
		{
			"type": "socket",

			// Socket type. Must be one of ["tcp", "unix", "unixpacket"].
      "socket": "tcp",

      // For TCP, address must have the form `host:port`.
      // For Unix networks, the address must be a file system path.
      "address": "localhost:9999"
		}
	]
}
```

> Note: at the moment, this input only works with connection-oriented sockets.
>
> Datagram-oriented sockets like UDP or UNIXGRAM socket are not supported.
