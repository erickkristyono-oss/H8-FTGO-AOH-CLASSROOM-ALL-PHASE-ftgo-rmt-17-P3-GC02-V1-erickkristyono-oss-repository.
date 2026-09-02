package usecase

import (
	"context"
	"time"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// Port (interface) berikut dideklarasikan di layer usecase sesuai prinsip
// Dependency Inversion: usecase mendefinisikan kebutuhannya, implementasi
// konkret (in-memory / postgres) berada di layer luar.

// UserRepository menangani persistensi User.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

// BookFilter dipakai untuk memfilter daftar buku.
type BookFilter struct {
	Status  domain.BookStatus // kosong = semua status
	OwnerID string            // kosong = semua pemilik
}

// BookRepository menangani persistensi Book.
type BookRepository interface {
	Create(ctx context.Context, b *domain.Book) error
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	List(ctx context.Context, f BookFilter) ([]*domain.Book, error)
	Update(ctx context.Context, b *domain.Book) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status domain.BookStatus) error
}

// BorrowRepository menangani persistensi BorrowedBook.
type BorrowRepository interface {
	Create(ctx context.Context, b *domain.BorrowedBook) error
	GetActiveByBook(ctx context.Context, bookID string) (*domain.BorrowedBook, error)
	FindActiveOverdue(ctx context.Context, now time.Time) ([]*domain.BorrowedBook, error)
	Update(ctx context.Context, b *domain.BorrowedBook) error
}

// PasswordHasher adalah kontrak hashing password (diimplementasikan security.Hasher).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) bool
}

// TokenGenerator adalah kontrak pembuat token (diimplementasikan token.Manager).
type TokenGenerator interface {
	Generate(userID, username string) (string, error)
}
