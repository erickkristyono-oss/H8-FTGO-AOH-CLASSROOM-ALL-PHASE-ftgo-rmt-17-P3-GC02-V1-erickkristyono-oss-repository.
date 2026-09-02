package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/repository/memory"
	"github.com/yourusername/book-management-grpc/internal/security"
	"github.com/yourusername/book-management-grpc/internal/token"
	"github.com/yourusername/book-management-grpc/internal/usecase"
)

func newAuthUsecase() *usecase.AuthUsecase {
	return usecase.NewAuthUsecase(
		memory.NewUserRepository(),
		security.NewHasher(),
		token.NewManager("test-secret", time.Hour),
	)
}

func TestRegister_Success(t *testing.T) {
	uc := newAuthUsecase()
	u, err := uc.Register(context.Background(), "alice", "password1")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if u.ID == "" {
		t.Error("user ID kosong")
	}
	if u.Password == "password1" {
		t.Error("password harus disimpan dalam bentuk hash")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	uc := newAuthUsecase()
	ctx := context.Background()
	if _, err := uc.Register(ctx, "bob", "password1"); err != nil {
		t.Fatalf("register pertama error: %v", err)
	}
	_, err := uc.Register(ctx, "bob", "password2")
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("error = %v, want ErrUserAlreadyExists", err)
	}
}

func TestRegister_ValidationError(t *testing.T) {
	uc := newAuthUsecase()
	if _, err := uc.Register(context.Background(), "x", "123"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation (password terlalu pendek)", err)
	}
}

func TestLogin_Success(t *testing.T) {
	uc := newAuthUsecase()
	ctx := context.Background()
	if _, err := uc.Register(ctx, "carol", "password1"); err != nil {
		t.Fatalf("register error: %v", err)
	}
	tok, u, err := uc.Login(ctx, "carol", "password1")
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if tok == "" {
		t.Error("token kosong")
	}
	if u.Username != "carol" {
		t.Errorf("username = %q, want carol", u.Username)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	uc := newAuthUsecase()
	ctx := context.Background()
	_, _ = uc.Register(ctx, "dave", "password1")
	_, _, err := uc.Login(ctx, "dave", "salahpassword")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	uc := newAuthUsecase()
	_, _, err := uc.Login(context.Background(), "ghost", "password1")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}
