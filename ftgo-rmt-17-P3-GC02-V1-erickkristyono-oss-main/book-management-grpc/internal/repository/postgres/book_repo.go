package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/usecase"
)

// BookRepository adalah implementasi Postgres untuk usecase.BookRepository.
type BookRepository struct{ db *gorm.DB }

// NewBookRepository membuat BookRepository.
func NewBookRepository(db *gorm.DB) *BookRepository { return &BookRepository{db: db} }

func (r *BookRepository) Create(ctx context.Context, b *domain.Book) error {
	m := &BookModel{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		PublishedDate: b.PublishedDate,
		Status:        string(b.Status),
		OwnerID:       b.OwnerID,
		UserID:        b.UserID,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var m BookModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainBook(&m), nil
}

func (r *BookRepository) List(ctx context.Context, f usecase.BookFilter) ([]*domain.Book, error) {
	q := r.db.WithContext(ctx).Model(&BookModel{})
	if f.Status != "" {
		q = q.Where("status = ?", string(f.Status))
	}
	if f.OwnerID != "" {
		q = q.Where("owner_id = ?", f.OwnerID)
	}
	var models []BookModel
	if err := q.Order("title asc").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Book, 0, len(models))
	for i := range models {
		out = append(out, toDomainBook(&models[i]))
	}
	return out, nil
}

func (r *BookRepository) Update(ctx context.Context, b *domain.Book) error {
	m := &BookModel{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		PublishedDate: b.PublishedDate,
		Status:        string(b.Status),
		OwnerID:       b.OwnerID,
		UserID:        b.UserID,
	}
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&BookModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BookRepository) UpdateStatus(ctx context.Context, id string, status domain.BookStatus) error {
	res := r.db.WithContext(ctx).Model(&BookModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
