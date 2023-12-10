gen:
	protoc -I. --go_out=internal/gen --go_opt=paths=source_relative \
    --go-grpc_out=internal/gen --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=internal/gen --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=swagger \
	proto/pcbook/v1/*.proto \

clean:
	rm -r internal/gen/proto/*
	rm swagger/proto/pcbook/**/*

server1:
	go run cmd/server/main.go -port 50051 

server2:
	go run cmd/server/main.go -port 50052

server1-tls:
	go run cmd/server/main.go -port 50051 -tls

server2-tls:
	go run cmd/server/main.go -port 50052 -tls

server:
	go run cmd/server/main.go -port 8080

rest:
	go run cmd/server/main.go -port 8081 -type rest -endpoint 0.0.0.0:8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080 

server-tls:
	go run cmd/server/main.go -port 8080 -tls

client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls

test:
	go test -cover -race ./...

cert:
	cd cert; ./gen.sh; cd ..

.PHONY: gen clean server client test cert