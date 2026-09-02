.PHONY: help proto-deps generate tidy test test-core cover run build docker-build docker-up docker-down clean

help: ## Tampilkan bantuan
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

proto-deps: ## Install buf & plugin protoc (sekali saja)
	go install github.com/bufbuild/buf/cmd/buf@v1.30.0

generate: ## Generate kode gRPC + gateway + swagger dari file .proto
	buf generate proto

tidy: ## Rapikan & unduh dependency
	go mod tidy

test: ## Jalankan SEMUA test
	go test ./... -cover

test-core: ## Jalankan test layer inti (tanpa perlu codegen)
	go test ./internal/domain/... ./internal/idgen/... ./internal/security/... \
		./internal/token/... ./internal/repository/memory/... \
		./internal/usecase/... ./internal/scheduler/... -v -cover

cover: ## Laporan coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

run: ## Jalankan server (mode in-memory)
	DB_DRIVER=memory go run ./cmd/server

build: ## Build binary ke ./server
	CGO_ENABLED=0 go build -o server ./cmd/server

docker-build: ## Build Docker image
	docker build -t book-management-grpc:latest .

docker-up: ## Jalankan via docker-compose (app + postgres)
	docker compose up --build

docker-down: ## Hentikan docker-compose
	docker compose down

clean: ## Bersihkan artefak
	rm -f server coverage.out
