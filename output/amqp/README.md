gogstash output amqp
=======================

## Synopsis

```
{
	"output": [
		{
			"type": "amqp",

			// Array of AMQP connection strings formatted per the [RabbitMQ URI Spec](http://www.rabbitmq.com/uri-spec.html). Required.
			// Note: the connections will be used in a round-robin fashions
			"urls": [
				"amqp://guest:guest@localhost:5672//vhost"
			],

			// The message routing key used to bind the queue to the exchange. Defaults to empty string.
			"routing_key": "%{fieldstring}"

			// AMQP exchange name. Required.
			"exchange": "amq.topic",

			// AMQP exchange type (fanout, direct, topic or headers).
			"exchange_type": "topic",

			// Whether the exchange should be configured as a durable exchange. Defaults to false.
			"exchange_durable": false,

			// Whether the exchange is deleted when all queues have finished and there is no publishing. Defaults to false.
			"exchange_auto_delete": false,

			// Whether published messages should be marked as persistent or transient. Defaults to false.
			"persistent": false,

			// Number of attempts for sending a message. Defaults to 3.
			"retries": 3,

			// Delay between each attempt to reconnect to AMQP server. Defaults to 30 seconds.
			"reconnect_delay": 30
		}
	]
}
```
