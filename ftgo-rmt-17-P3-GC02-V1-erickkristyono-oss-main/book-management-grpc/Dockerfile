# ---- Stage 1: build ---------------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependency (butuh go.sum -> jalankan `go mod tidy` sebelum build).
COPY go.mod go.sum ./
RUN go mod download

# Salin sisa source (termasuk pkg/genproto hasil `make generate`).
COPY . .

# Build binary statis.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ---- Stage 2: runtime -------------------------------------------------------
FROM alpine:3.19

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /server /app/server
COPY --from=builder /app/docs/swagger /app/docs/swagger

# gRPC 50051, HTTP/Swagger 8080
EXPOSE 50051 8080

ENV GRPC_PORT=50051 \
    HTTP_PORT=8080 \
    DB_DRIVER=memory

ENTRYPOINT ["/app/server"]
