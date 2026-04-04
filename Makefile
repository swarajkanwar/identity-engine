.PHONY: install-tools generate deps build run clean all

all: install-tools generate deps build

install-tools:
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install protobuf || true; \
	elif [ "$$(uname)" = "Linux" ]; then \
		sudo apt-get install -y protobuf-compiler || true; \
	else \
		echo "Please install protoc manually on this OS"; \
	fi
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	PATH="$${PATH}:$$(go env GOPATH)/bin" protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/identity.proto

deps:
	go get -u google.golang.org/grpc
	go get -u google.golang.org/protobuf
	go mod tidy

build:
	go build -o bin/identity-engine internal/cmd/main.go
	go build -o bin/identity-mcp internal/mcp/main.go

run: build
	./bin/identity-engine

run-mcp: build
	./bin/identity-mcp

clean:
	rm -rf bin/
	rm -f proto/*.pb.go
