package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// UserRepository adalah implementasi Postgres untuk usecase.UserRepository.
type UserRepository struct{ db *gorm.DB }

// NewUserRepository membuat UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	m := &UserModel{ID: u.ID, Username: u.Username, Password: u.Password}
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil && errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrUserAlreadyExists
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(&m), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(&m), nil
}
