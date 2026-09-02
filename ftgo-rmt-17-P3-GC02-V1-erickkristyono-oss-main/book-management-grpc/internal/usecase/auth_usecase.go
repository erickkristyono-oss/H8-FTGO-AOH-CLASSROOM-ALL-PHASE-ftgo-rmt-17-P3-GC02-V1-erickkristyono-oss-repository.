package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/idgen"
)

// AuthUsecase menangani registrasi & login pengguna.
type AuthUsecase struct {
	users  UserRepository
	hasher PasswordHasher
	tokens TokenGenerator
}

// NewAuthUsecase membuat AuthUsecase.
func NewAuthUsecase(users UserRepository, hasher PasswordHasher, tokens TokenGenerator) *AuthUsecase {
	return &AuthUsecase{users: users, hasher: hasher, tokens: tokens}
}

// Register mendaftarkan user baru dan mengembalikan user yang tersimpan.
func (uc *AuthUsecase) Register(ctx context.Context, username, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(password) < 6 {
		return nil, errors.Join(domain.ErrValidation,
			errors.New("username wajib diisi dan password minimal 6 karakter"))
	}

	// Pastikan username belum dipakai.
	if _, err := uc.users.GetByUsername(ctx, username); err == nil {
		return nil, domain.ErrUserAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	hashed, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		ID:       idgen.NewUUID(),
		Username: username,
		Password: hashed,
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login memverifikasi kredensial dan mengembalikan JWT token.
func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
	username = strings.TrimSpace(username)
	u, err := uc.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", nil, domain.ErrInvalidCredentials
		}
		return "", nil, err
	}
	if !uc.hasher.Verify(password, u.Password) {
		return "", nil, domain.ErrInvalidCredentials
	}
	tok, err := uc.tokens.Generate(u.ID, u.Username)
	if err != nil {
		return "", nil, err
	}
	return tok, u, nil
}
