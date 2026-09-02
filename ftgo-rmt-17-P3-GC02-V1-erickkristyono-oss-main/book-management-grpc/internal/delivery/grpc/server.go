package grpcdelivery

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	bookv1 "github.com/yourusername/book-management-grpc/pkg/genproto/book/v1"
)

// NewGRPCServer membangun *grpc.Server, memasang interceptor JWT, dan
// mendaftarkan seluruh service. Reflection diaktifkan agar bisa diuji
// dengan grpcurl / Postman.
func NewGRPCServer(authH *AuthHandler, bookH *BookHandler, interceptor *AuthInterceptor) *grpc.Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.Unary()),
	)
	bookv1.RegisterAuthServiceServer(s, authH)
	bookv1.RegisterBookServiceServer(s, bookH)
	reflection.Register(s)
	return s
}
