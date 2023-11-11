gen:
	protoc --go_out=internal/gen --go_opt=paths=source_relative \
    --go-grpc_out=internal/gen --go-grpc_opt=paths=source_relative \
	proto/pcbook/v1/*.proto

clean:
	rm -r internal/gen/proto/*

run:
	go run main.go