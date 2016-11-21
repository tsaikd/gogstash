gogstash output email
=======================

## Synopsis

```
{
    "output": [
        {
            "type": "email",

            // (required)
            "address": "smtp.xxx.com",

            // (required)
            "username": "your_user_name",

            // (required)
            "password": "your_password",

            // (required)
            "subject": "your subject",

            // (required)
            "from": "from@youremail",

            // (required)
            "to": "your1@youremail.com;your2@youremail.com",

            // (optional)
            "cc": "cc1@youremail.com;cc2@youremail.com",

            // (optional)
            "port": 25,

            // (optional)
            "use_tls": true,
        }
    ]
}
```