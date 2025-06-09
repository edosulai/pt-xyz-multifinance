.PHONY: build test migrate-up migrate-down proto run check-proto-tools

check-proto-tools:
	@which protoc > /dev/null || (echo "protoc is not installed" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "protoc-gen-go is not installed" && exit 1)
	@which protoc-gen-go-grpc > /dev/null || (echo "protoc-gen-go-grpc is not installed" && exit 1)
	@which protoc-gen-grpc-gateway > /dev/null || (echo "protoc-gen-grpc-gateway is not installed" && exit 1)
	@which protoc-gen-openapiv2 > /dev/null || (echo "protoc-gen-openapiv2 is not installed" && exit 1)

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database "postgresql://postgres:root@localhost:5432/xyz_multifinance?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:root@localhost:5432/xyz_multifinance?sslmode=disable" down

proto: check-proto-tools
	mkdir -p proto/gen/openapiv2
	protoc -I . \
		-I third_party \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		--openapiv2_out proto/gen/openapiv2 --openapiv2_opt logtostderr=true \
		$(shell find proto -name "*.proto" -not -path "*/third_party/*")

run:
	go run ./cmd
