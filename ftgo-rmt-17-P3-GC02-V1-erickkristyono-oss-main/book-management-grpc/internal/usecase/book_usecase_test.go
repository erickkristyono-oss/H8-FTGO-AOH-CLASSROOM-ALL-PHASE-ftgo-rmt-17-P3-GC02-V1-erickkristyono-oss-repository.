package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/repository/memory"
	"github.com/yourusername/book-management-grpc/internal/usecase"
)

func newBookUsecase() *usecase.BookUsecase {
	return usecase.NewBookUsecase(memory.NewBookRepository(), memory.NewBorrowRepository())
}

func TestCreateBook(t *testing.T) {
	uc := newBookUsecase()
	b, err := uc.CreateBook(context.Background(), usecase.CreateBookInput{
		Title:   "Clean Architecture",
		Author:  "Robert C. Martin",
		OwnerID: "owner-1",
	})
	if err != nil {
		t.Fatalf("CreateBook error: %v", err)
	}
	if b.Status != domain.StatusAvailable {
		t.Errorf("status = %q, want Available", b.Status)
	}
	if b.OwnerID != "owner-1" {
		t.Errorf("ownerID = %q, want owner-1", b.OwnerID)
	}
}

func TestCreateBook_Validation(t *testing.T) {
	uc := newBookUsecase()
	_, err := uc.CreateBook(context.Background(), usecase.CreateBookInput{Title: "", Author: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}

func TestUpdateBook_ForbiddenForNonOwner(t *testing.T) {
	uc := newBookUsecase()
	ctx := context.Background()
	b, _ := uc.CreateBook(ctx, usecase.CreateBookInput{Title: "Go", Author: "A", OwnerID: "owner-1"})

	_, err := uc.UpdateBook(ctx, usecase.UpdateBookInput{
		ID:      b.ID,
		Title:   "Hacked",
		ActorID: "attacker",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestBorrowAndReturnBook(t *testing.T) {
	uc := newBookUsecase()
	ctx := context.Background()
	b, _ := uc.CreateBook(ctx, usecase.CreateBookInput{Title: "DDD", Author: "Evans", OwnerID: "owner-1"})

	// Pinjam.
	rec, err := uc.BorrowBook(ctx, b.ID, "borrower-1", 24*time.Hour)
	if err != nil {
		t.Fatalf("BorrowBook error: %v", err)
	}
	if rec.ReturnDate != nil {
		t.Error("ReturnDate harus nil setelah dipinjam")
	}
	got, _ := uc.GetBook(ctx, b.ID)
	if got.Status != domain.StatusBorrowed {
		t.Errorf("status = %q, want Borrowed", got.Status)
	}

	// Pinjam lagi buku yang sama -> harus gagal (tidak available).
	if _, err := uc.BorrowBook(ctx, b.ID, "borrower-2", 0); !errors.Is(err, domain.ErrBookNotAvailable) {
		t.Fatalf("error = %v, want ErrBookNotAvailable", err)
	}

	// Kembalikan.
	if err := uc.ReturnBook(ctx, b.ID, "borrower-1"); err != nil {
		t.Fatalf("ReturnBook error: %v", err)
	}
	got, _ = uc.GetBook(ctx, b.ID)
	if got.Status != domain.StatusAvailable {
		t.Errorf("status setelah return = %q, want Available", got.Status)
	}
}

func TestReturnBook_ForbiddenForNonBorrower(t *testing.T) {
	uc := newBookUsecase()
	ctx := context.Background()
	b, _ := uc.CreateBook(ctx, usecase.CreateBookInput{Title: "X", Author: "Y", OwnerID: "o"})
	_, _ = uc.BorrowBook(ctx, b.ID, "borrower-1", time.Hour)

	if err := uc.ReturnBook(ctx, b.ID, "orang-lain"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

// TestMarkOverdueBooks menguji logika inti job scheduler.
func TestMarkOverdueBooks(t *testing.T) {
	uc := newBookUsecase()
	ctx := context.Background()

	// Gunakan jam palsu: "sekarang" = waktu tetap.
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	uc.WithClock(func() time.Time { return base })

	b, _ := uc.CreateBook(ctx, usecase.CreateBookInput{Title: "Late", Author: "A", OwnerID: "o"})
	// Pinjam dengan durasi 1 jam -> DueDate = base + 1h.
	if _, err := uc.BorrowBook(ctx, b.ID, "borrower-1", time.Hour); err != nil {
		t.Fatalf("borrow error: %v", err)
	}

	// Belum lewat tenggat: tidak ada yang overdue.
	n, err := uc.MarkOverdueBooks(ctx)
	if err != nil {
		t.Fatalf("MarkOverdueBooks error: %v", err)
	}
	if n != 0 {
		t.Fatalf("overdue = %d, want 0 (belum lewat tenggat)", n)
	}

	// Majukan waktu 2 jam -> sudah lewat tenggat.
	uc.WithClock(func() time.Time { return base.Add(2 * time.Hour) })
	n, err = uc.MarkOverdueBooks(ctx)
	if err != nil {
		t.Fatalf("MarkOverdueBooks error: %v", err)
	}
	if n != 1 {
		t.Fatalf("overdue = %d, want 1", n)
	}
	got, _ := uc.GetBook(ctx, b.ID)
	if got.Status != domain.StatusOverdue {
		t.Errorf("status = %q, want Overdue", got.Status)
	}

	// Idempotent: menjalankan lagi tidak menambah (sudah Overdue).
	n, _ = uc.MarkOverdueBooks(ctx)
	if n != 0 {
		t.Errorf("run kedua overdue = %d, want 0 (idempotent)", n)
	}
}
