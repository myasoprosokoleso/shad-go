PROTO_DIR := proto
GEN_DIR   := gen
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Use `docker compose` (v2 plugin) if available, else fall back to `docker-compose`.
DC := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

.PHONY: gen tools server client loadtest build clean up down logs ps

# Generate Go code from .proto files.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc (see README).
gen:
	@mkdir -p $(GEN_DIR)
	protoc \
		--go_out=.     --go_opt=module=playground/grpc-demo \
		--go-grpc_out=. --go-grpc_opt=module=playground/grpc-demo \
		--proto_path=$(PROTO_DIR) \
		$(PROTO_FILES)
	@echo "generated -> $(GEN_DIR)/"

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# --- host-side runs (server typically runs in docker, see `make up`) ---

server:
	go run ./server

client:
	go run ./client

# Example: make loadtest ARGS="-rps 500 -duration 5s -method mul -trace -trace-sample 0.05"
loadtest:
	go run ./loadtest $(ARGS)

build:
	go build -o bin/server   ./server
	go build -o bin/client   ./client
	go build -o bin/loadtest ./loadtest

# --- docker-compose ---

up:
	$(DC) up -d --build
	@echo
	@echo "  Jaeger UI:    http://localhost:16686"
	@echo "  gRPC server:  localhost:50051"
	@echo "  OTLP/gRPC:    localhost:4317  (for host-side client/loadtest)"

down:
	$(DC) down

logs:
	$(DC) logs -f --tail=100

ps:
	$(DC) ps

clean:
	rm -rf bin $(GEN_DIR)/calculatorpb/*.pb.go
