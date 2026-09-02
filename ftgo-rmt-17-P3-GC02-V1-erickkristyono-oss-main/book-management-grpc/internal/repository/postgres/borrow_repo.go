package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// BorrowRepository adalah implementasi Postgres untuk usecase.BorrowRepository.
type BorrowRepository struct{ db *gorm.DB }

// NewBorrowRepository membuat BorrowRepository.
func NewBorrowRepository(db *gorm.DB) *BorrowRepository { return &BorrowRepository{db: db} }

func (r *BorrowRepository) Create(ctx context.Context, b *domain.BorrowedBook) error {
	m := &BorrowedBookModel{
		ID:           b.ID,
		BookID:       b.BookID,
		UserID:       b.UserID,
		BorrowedDate: b.BorrowedDate,
		DueDate:      b.DueDate,
		ReturnDate:   b.ReturnDate,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *BorrowRepository) GetActiveByBook(ctx context.Context, bookID string) (*domain.BorrowedBook, error) {
	var m BorrowedBookModel
	err := r.db.WithContext(ctx).
		Where("book_id = ? AND return_date IS NULL", bookID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainBorrow(&m), nil
}

func (r *BorrowRepository) FindActiveOverdue(ctx context.Context, now time.Time) ([]*domain.BorrowedBook, error) {
	var models []BorrowedBookModel
	err := r.db.WithContext(ctx).
		Where("return_date IS NULL AND due_date < ?", now).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.BorrowedBook, 0, len(models))
	for i := range models {
		out = append(out, toDomainBorrow(&models[i]))
	}
	return out, nil
}

func (r *BorrowRepository) Update(ctx context.Context, b *domain.BorrowedBook) error {
	m := &BorrowedBookModel{
		ID:           b.ID,
		BookID:       b.BookID,
		UserID:       b.UserID,
		BorrowedDate: b.BorrowedDate,
		DueDate:      b.DueDate,
		ReturnDate:   b.ReturnDate,
	}
	return r.db.WithContext(ctx).Save(m).Error
}
