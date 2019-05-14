# arenawithfriends
REST API to create play sessions for MTG Arena

A Magic the Gathering tool to create card pools based on cards every player owns.

Example: 
* Player A owns 3 Chandras and 1 Jace
* Player B owns 2 Chandras and 2 Jaces

common card pool: 2 Chandras, 1 Jace.

## Installation
### local binary, RAM session cache
```
go get -d -u github.com/kjeisy/arenawithfriends
```

### Google AppEngine
In the folder `appengine` run (requires a Google Cloud account and the SDK):
```
gcloud app deploy
```

## API calls

- initiate a new session: empty POST to `/api/v1/session`
```
$ curl -d '{}' http://localhost:8080/api/v1/session
{"id":"pit7UzLlf8azujxZz7vL"}
```

- look at session (by ID): GET to `/api/v1/session/<ID>`
```
$ curl http://localhost:8080/api/v1/session/pit7UzLlf8azujxZz7vL
{"players":null,"started":false}
```

- add data for players: POST to `/api/v1/session/<ID>/player` (parameters name, collection)
```
$ curl -d '{"name":"test","collection":{"123":3,"345":2}}' http://localhost:8080/api/v1/session/pit7UzLlf8azujxZz7vL/player
{"players":["test"],"started":false}
```
```
$ curl -d '{"name":"test2","collection":{"124":3,"345":1}}' http://localhost:8080/api/v1/session/pit7UzLlf8azujxZz7vL/player
{"players":["test","test2"],"started":false}
```

- start session: POST to `/api/v1/session/<ID>/start`
```
$ curl -d '{}' http://localhost:8080/api/v1/session/pit7UzLlf8azujxZz7vL/start
{"players":["test","test2"],"started":true}
```

- get collection: GET to `/api/v1/session/<ID>/collection`
```
$ curl http://localhost:8080/api/v1/session/pit7UzLlf8azujxZz7vL/collection
{"345":1}
```
