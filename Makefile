.PHONY: build test migrate-up migrate-down proto run

build:
	go build -o bin/server.exe ./cmd

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database "postgresql://postgres:root@localhost:5432/xyz_multifinance?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:root@localhost:5432/xyz_multifinance?sslmode=disable" down

proto:
	protoc -I . \
		-I third_party \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		proto/user.proto

run:
	go run ./cmd
