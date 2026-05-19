# Brainway — Transaction Ingestion Service

A gRPC microservice that ingests financial transactions and queues them for async processing via [asynq](https://github.com/hibiken/asynq).

## Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [protoc](https://github.com/protocolbuffers/protobuf/releases) (Protocol Buffers compiler)

## Setup

Install the Go protoc plugins:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Install Go dependencies:

```sh
go mod tidy
```

## Generate gRPC code

Run from the project root whenever `proto/transaction.proto` changes:

```sh
protoc \
  --proto_path=proto \
  --go_out=pb --go_opt=paths=source_relative \
  --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
  proto/transaction.proto
```

On Windows (PowerShell):

```powershell
protoc `
  --proto_path=proto `
  --go_out=pb --go_opt=paths=source_relative `
  --go-grpc_out=pb --go-grpc_opt=paths=source_relative `
  proto/transaction.proto
```

This regenerates `pb/transaction.pb.go` and `pb/transaction_grpc.pb.go`.

## Run tests

```sh
go test ./internal/handler/... -v -count=1
```

To test all packages:

```sh
go test ./... -count=1
```

## Project structure

```
proto/              # Protobuf source definitions
pb/                 # Generated Go code (do not edit manually)
internal/
  queue/            # Enqueuer interface and mock for testing
  handler/          # IngestBatch handler and unit tests
cmd/server/         # Server entry point (connects to Redis on localhost:6379)
```

## Run the server

```sh
go run ./cmd/server
```
