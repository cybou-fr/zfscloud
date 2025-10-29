.PHONY: proto build test clean server client

# Protobuf generation
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# Build targets
build: server client

server:
	go build -o bin/zfscloud-server ./server/cmd

client:
	go build -o bin/zfscloud-client ./client/cmd

# Testing
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Development
run-server:
	go run ./server/cmd -config server/config.yaml

run-client:
	go run ./client/cmd -config client/config.yaml

# Linting
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install development dependencies
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
