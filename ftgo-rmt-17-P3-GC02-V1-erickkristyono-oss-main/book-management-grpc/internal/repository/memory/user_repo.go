// Package memory berisi implementasi repository berbasis in-memory (map).
// Cocok untuk unit test dan menjalankan aplikasi tanpa database eksternal.
package memory

import (
	"context"
	"sync"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// UserRepository menyimpan user di memori.
type UserRepository struct {
	mu     sync.RWMutex
	byID   map[string]*domain.User
	byName map[string]*domain.User
}

// NewUserRepository membuat repository user in-memory kosong.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:   make(map[string]*domain.User),
		byName: make(map[string]*domain.User),
	}
}

func (r *UserRepository) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byName[u.Username]; ok {
		return domain.ErrUserAlreadyExists
	}
	cp := *u
	r.byID[u.ID] = &cp
	r.byName[u.Username] = &cp
	return nil
}

func (r *UserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepository) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byName[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
