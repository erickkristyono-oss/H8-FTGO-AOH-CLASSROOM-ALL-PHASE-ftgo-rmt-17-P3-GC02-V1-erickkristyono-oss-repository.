package grpcdelivery

import (
	"context"

	"github.com/yourusername/book-management-grpc/internal/token"
	"github.com/yourusername/book-management-grpc/internal/usecase"
	bookv1 "github.com/yourusername/book-management-grpc/pkg/genproto/book/v1"
)

// AuthHandler mengimplementasikan bookv1.AuthServiceServer.
type AuthHandler struct {
	bookv1.UnimplementedAuthServiceServer
	auth   *usecase.AuthUsecase
	tokens *token.Manager
}

// NewAuthHandler membuat AuthHandler.
func NewAuthHandler(auth *usecase.AuthUsecase, tokens *token.Manager) *AuthHandler {
	return &AuthHandler{auth: auth, tokens: tokens}
}

// Register mendaftarkan user baru dan langsung mengembalikan token (auto-login).
func (h *AuthHandler) Register(ctx context.Context, req *bookv1.RegisterRequest) (*bookv1.AuthResponse, error) {
	u, err := h.auth.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}
	tok, err := h.tokens.Generate(u.ID, u.Username)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &bookv1.AuthResponse{
		Token: tok,
		User:  &bookv1.User{Id: u.ID, Username: u.Username},
	}, nil
}

// Login memverifikasi kredensial & mengembalikan token JWT.
func (h *AuthHandler) Login(ctx context.Context, req *bookv1.LoginRequest) (*bookv1.AuthResponse, error) {
	tok, u, err := h.auth.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &bookv1.AuthResponse{
		Token: tok,
		User:  &bookv1.User{Id: u.ID, Username: u.Username},
	}, nil
}
