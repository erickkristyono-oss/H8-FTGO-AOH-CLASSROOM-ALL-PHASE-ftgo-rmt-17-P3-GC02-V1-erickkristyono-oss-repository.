package grpcdelivery

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bookv1 "github.com/yourusername/book-management-grpc/pkg/genproto/book/v1"
)

// NewGatewayHandler membangun http.Handler yang:
//   - mem-proxy REST -> gRPC di bawah prefix /v1/ (grpc-gateway),
//   - menyajikan Swagger UI di /swagger/,
//   - menyediakan health check di /healthz.
//
// grpcEndpoint contoh: "localhost:50051". swaggerDir adalah folder berisi
// index.html + openapi.json.
func NewGatewayHandler(ctx context.Context, grpcEndpoint, swaggerDir string) (http.Handler, error) {
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := bookv1.RegisterAuthServiceHandlerFromEndpoint(ctx, gwMux, grpcEndpoint, opts); err != nil {
		return nil, err
	}
	if err := bookv1.RegisterBookServiceHandlerFromEndpoint(ctx, gwMux, grpcEndpoint, opts); err != nil {
		return nil, err
	}

	root := http.NewServeMux()
	root.Handle("/v1/", gwMux)
	root.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(swaggerDir))))
	root.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return root, nil
}
