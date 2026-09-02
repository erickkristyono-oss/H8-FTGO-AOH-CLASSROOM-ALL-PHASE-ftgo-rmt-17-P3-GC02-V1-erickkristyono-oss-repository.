package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/idgen"
)

// DefaultBorrowDuration adalah lama peminjaman default bila tidak dispesifikasikan.
const DefaultBorrowDuration = 7 * 24 * time.Hour

// BookUsecase menangani manajemen buku & peminjaman.
type BookUsecase struct {
	books   BookRepository
	borrows BorrowRepository
	now     func() time.Time // di-inject agar mudah ditest
}

// NewBookUsecase membuat BookUsecase.
func NewBookUsecase(books BookRepository, borrows BorrowRepository) *BookUsecase {
	return &BookUsecase{books: books, borrows: borrows, now: time.Now}
}

// WithClock mengganti sumber waktu (dipakai pada test).
func (uc *BookUsecase) WithClock(now func() time.Time) *BookUsecase {
	uc.now = now
	return uc
}

// CreateBookInput adalah data pembuatan buku.
type CreateBookInput struct {
	Title         string
	Author        string
	PublishedDate time.Time
	OwnerID       string // diambil dari user yang terautentikasi
}

// CreateBook menambah buku baru milik OwnerID dengan status Available.
func (uc *BookUsecase) CreateBook(ctx context.Context, in CreateBookInput) (*domain.Book, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Author) == "" {
		return nil, errors.Join(domain.ErrValidation, errors.New("title & author wajib diisi"))
	}
	b := &domain.Book{
		ID:            idgen.NewUUID(),
		Title:         in.Title,
		Author:        in.Author,
		PublishedDate: in.PublishedDate,
		Status:        domain.StatusAvailable,
		OwnerID:       in.OwnerID,
	}
	if err := uc.books.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBook mengambil satu buku berdasarkan id.
func (uc *BookUsecase) GetBook(ctx context.Context, id string) (*domain.Book, error) {
	return uc.books.GetByID(ctx, id)
}

// ListBooks mengambil daftar buku sesuai filter.
func (uc *BookUsecase) ListBooks(ctx context.Context, f BookFilter) ([]*domain.Book, error) {
	return uc.books.List(ctx, f)
}

// UpdateBookInput adalah data perubahan buku.
type UpdateBookInput struct {
	ID            string
	Title         string
	Author        string
	PublishedDate time.Time
	ActorID       string // user yang melakukan aksi (untuk otorisasi)
}

// UpdateBook memperbarui data buku. Hanya pemilik (OwnerID) yang boleh mengubah.
func (uc *BookUsecase) UpdateBook(ctx context.Context, in UpdateBookInput) (*domain.Book, error) {
	b, err := uc.books.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if b.OwnerID != in.ActorID {
		return nil, domain.ErrForbidden
	}
	if strings.TrimSpace(in.Title) != "" {
		b.Title = in.Title
	}
	if strings.TrimSpace(in.Author) != "" {
		b.Author = in.Author
	}
	if !in.PublishedDate.IsZero() {
		b.PublishedDate = in.PublishedDate
	}
	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// DeleteBook menghapus buku. Hanya pemilik yang boleh menghapus.
func (uc *BookUsecase) DeleteBook(ctx context.Context, id, actorID string) error {
	b, err := uc.books.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if b.OwnerID != actorID {
		return domain.ErrForbidden
	}
	return uc.books.Delete(ctx, id)
}

// BorrowBook meminjam buku untuk userID. Buku harus berstatus Available.
func (uc *BookUsecase) BorrowBook(ctx context.Context, bookID, userID string, duration time.Duration) (*domain.BorrowedBook, error) {
	if duration <= 0 {
		duration = DefaultBorrowDuration
	}
	b, err := uc.books.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if b.Status != domain.StatusAvailable {
		return nil, domain.ErrBookNotAvailable
	}

	now := uc.now()
	record := &domain.BorrowedBook{
		ID:           idgen.NewUUID(),
		BookID:       bookID,
		UserID:       userID,
		BorrowedDate: now,
		DueDate:      now.Add(duration),
		ReturnDate:   nil,
	}
	if err := uc.borrows.Create(ctx, record); err != nil {
		return nil, err
	}

	b.Status = domain.StatusBorrowed
	b.UserID = userID
	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	return record, nil
}

// ReturnBook mengembalikan buku yang sedang dipinjam userID.
func (uc *BookUsecase) ReturnBook(ctx context.Context, bookID, userID string) error {
	record, err := uc.borrows.GetActiveByBook(ctx, bookID)
	if err != nil {
		return err
	}
	if record.UserID != userID {
		return domain.ErrForbidden
	}
	now := uc.now()
	record.ReturnDate = &now
	if err := uc.borrows.Update(ctx, record); err != nil {
		return err
	}
	return uc.books.UpdateStatus(ctx, bookID, domain.StatusAvailable)
}

// MarkOverdueBooks adalah pekerjaan yang dijalankan job scheduler:
// menandai buku yang peminjamannya lewat tenggat menjadi status Overdue.
// Mengembalikan jumlah buku yang diperbarui.
func (uc *BookUsecase) MarkOverdueBooks(ctx context.Context) (int, error) {
	now := uc.now()
	overdue, err := uc.borrows.FindActiveOverdue(ctx, now)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, rec := range overdue {
		book, err := uc.books.GetByID(ctx, rec.BookID)
		if err != nil {
			continue // buku mungkin sudah dihapus; lewati
		}
		if book.Status == domain.StatusOverdue {
			continue // sudah ditandai sebelumnya
		}
		if err := uc.books.UpdateStatus(ctx, rec.BookID, domain.StatusOverdue); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
