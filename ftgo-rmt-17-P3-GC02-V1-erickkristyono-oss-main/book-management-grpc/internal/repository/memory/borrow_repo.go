package memory

import (
	"context"
	"sync"
	"time"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// BorrowRepository menyimpan catatan peminjaman di memori.
type BorrowRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.BorrowedBook
}

// NewBorrowRepository membuat repository peminjaman in-memory kosong.
func NewBorrowRepository() *BorrowRepository {
	return &BorrowRepository{data: make(map[string]*domain.BorrowedBook)}
}

func (r *BorrowRepository) Create(_ context.Context, b *domain.BorrowedBook) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *b
	r.data[b.ID] = &cp
	return nil
}

// GetActiveByBook mengembalikan peminjaman aktif (belum dikembalikan) untuk suatu buku.
func (r *BorrowRepository) GetActiveByBook(_ context.Context, bookID string) (*domain.BorrowedBook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.data {
		if b.BookID == bookID && !b.IsReturned() {
			cp := *b
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

// FindActiveOverdue mengembalikan seluruh peminjaman yang belum dikembalikan
// dan sudah melewati DueDate pada waktu now.
func (r *BorrowRepository) FindActiveOverdue(_ context.Context, now time.Time) ([]*domain.BorrowedBook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.BorrowedBook, 0)
	for _, b := range r.data {
		if b.IsOverdue(now) {
			cp := *b
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *BorrowRepository) Update(_ context.Context, b *domain.BorrowedBook) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[b.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *b
	r.data[b.ID] = &cp
	return nil
}
