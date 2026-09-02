// Package grpcdelivery berisi lapisan transport gRPC: interceptor autentikasi,
// pemetaan error, dan handler yang mengimplementasikan service dari file proto.
package grpcdelivery

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yourusername/book-management-grpc/internal/token"
)

type ctxKey string

const userIDKey ctxKey = "userID"

// UserIDFromContext mengambil user id yang disisipkan oleh interceptor.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// AuthInterceptor memvalidasi JWT untuk method yang tidak publik.
type AuthInterceptor struct {
	tokens *token.Manager
	public map[string]bool
}

// NewAuthInterceptor membuat interceptor. Method Register & Login bersifat publik.
func NewAuthInterceptor(t *token.Manager) *AuthInterceptor {
	return &AuthInterceptor{
		tokens: t,
		public: map[string]bool{
			"/book.v1.AuthService/Register": true,
			"/book.v1.AuthService/Login":    true,
		},
	}
}

// Unary mengembalikan grpc.UnaryServerInterceptor.
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if a.public[info.FullMethod] {
			return handler(ctx, req)
		}
		userID, err := a.authorize(ctx)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, userIDKey, userID)
		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata tidak ditemukan")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "token authorization tidak ada")
	}
	raw := strings.TrimSpace(values[0])
	// Buang prefix "Bearer " bila ada.
	if i := strings.IndexByte(raw, ' '); i > 0 && strings.EqualFold(raw[:i], "bearer") {
		raw = strings.TrimSpace(raw[i+1:])
	}
	claims, err := a.tokens.Parse(raw)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "token tidak valid: "+err.Error())
	}
	return claims.Sub, nil
}
