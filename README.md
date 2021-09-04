# emag-homework

## How to run

Start db controller and nodes:

```
make db
make dbnode1
make dbnode2
```

Start the app and the client:

```
make app
make client
```

## Implementation

Tried DynamoDB style approach (1 coordinator, N nodes, replication on read) but it is what it is. :)