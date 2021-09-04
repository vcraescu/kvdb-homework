.PHONY: proto
proto:
	protoc internal/app/api/v1/*.proto \
		--go_out=plugins=grpc:. \
		--proto_path=.

	protoc internal/db/api/v1/*.proto \
		--go_out=plugins=grpc:. \
		--proto_path=.

.PHONY: app
app:
	@APP_ADDRESS=0.0.0.0:9000 DB_ADDRESS=0.0.0.0:8000 go run cmd/app/main.go

.PHONY: client
client:
	@APP_ADDRESS=0.0.0.0:9000 go run cmd/client/main.go

.PHONY: db
db:
	@CTRL_ADDRESS=0.0.0.0:8000 go run cmd/db/main.go

.PHONY: dbnode1
dbnode1:
	@CTRL_ADDRESS=0.0.0.0:8000 NODE_ADDRESS=0.0.0.0:0 STORE_PATH=/tmp/node1.json go run cmd/db/main.go -node

.PHONY: dbnode2
dbnode2:
	@CTRL_ADDRESS=0.0.0.0:8000 NODE_ADDRESS=0.0.0.0:0 STORE_PATH=/tmp/node2.json go run cmd/db/main.go -node

.PHONY: test
test:
	go test -race ./...

.PHONY: vet
vet:
	go vet ./...
