gen:
	protoc --go_out=internal/gen --go_opt=paths=source_relative \
    --go-grpc_out=internal/gen --go-grpc_opt=paths=source_relative \
	proto/pcbook/v1/*.proto

clean:
	rm -r internal/gen/proto/*

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

test:
	go test -cover -race ./...

.PHONY: gen clean server client test