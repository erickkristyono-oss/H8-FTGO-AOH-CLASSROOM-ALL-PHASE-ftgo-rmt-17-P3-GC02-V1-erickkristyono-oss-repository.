package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/usecase"
)

// BookRepository menyimpan buku di memori.
type BookRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Book
}

// NewBookRepository membuat repository buku in-memory kosong.
func NewBookRepository() *BookRepository {
	return &BookRepository{data: make(map[string]*domain.Book)}
}

func (r *BookRepository) Create(_ context.Context, b *domain.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *b
	r.data[b.ID] = &cp
	return nil
}

func (r *BookRepository) GetByID(_ context.Context, id string) (*domain.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *BookRepository) List(_ context.Context, f usecase.BookFilter) ([]*domain.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Book, 0, len(r.data))
	for _, b := range r.data {
		if f.Status != "" && b.Status != f.Status {
			continue
		}
		if f.OwnerID != "" && b.OwnerID != f.OwnerID {
			continue
		}
		cp := *b
		out = append(out, &cp)
	}
	// Urutkan agar deterministik (memudahkan test & tampilan).
	sort.Slice(out, func(i, j int) bool { return out[i].Title < out[j].Title })
	return out, nil
}

func (r *BookRepository) Update(_ context.Context, b *domain.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[b.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *b
	r.data[b.ID] = &cp
	return nil
}

func (r *BookRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

func (r *BookRepository) UpdateStatus(_ context.Context, id string, status domain.BookStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.data[id]
	if !ok {
		return domain.ErrNotFound
	}
	b.Status = status
	return nil
}
