gogstash output MongoDB
======================

## Synopsis

##### JSON Configuration file
```json
{
  "output": [
    {
      "type": "mongodb",
      "host": [
        "localhost:27017"
      ],
      "database": "gogstash",
      "collection": "allLogs",
      "timeout": 10,
      "connections": 10,
      "username": "username",
      "password": "password",
      "mechanism": "SCRAM-SHA-1",
      "retry_interval": 10
    }
  ]
}
```

##### Yaml configuration file
```yaml
---
output:
- type: mongodb
  host:
  - localhost:27017
  database: gogstash
  collection: allLogs
  timeout: 10
  connections: 10
  username: username
  password: password
  mechanism: SCRAM-SHA-1
  retry_interval: 10
```

## Details

* type: Must be **"mongodb"**
* host: The hostname(s) and port(s) of your MongoDB server(s) (default: localhost:27717)
* database: The name of the database (default: gogstash)
* collection: The name of the collection (default: allLogs)
* timeout: MongoDB initial connection timeout in seconds (default: 10)
* connections: The number of connections in the pool (default: 10)
* username: the username (default: username)
* password: the password (default: password)
* mechanism: the authentication mechanisms -- SCRAM-SHA-1, MONGODB-CR (default: MONGODB-CR)
* retry_interval: Interval for reconnecting to failed MongoDB connections, in seconds (default: 10)

