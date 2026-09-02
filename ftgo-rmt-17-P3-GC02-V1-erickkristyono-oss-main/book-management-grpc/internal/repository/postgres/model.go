// Package postgres berisi implementasi repository berbasis PostgreSQL (GORM).
// Dipakai pada mode produksi (DB_DRIVER=postgres).
package postgres

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourusername/book-management-grpc/internal/domain"
)

// ---- GORM models (representasi tabel) -------

// UserModel -> tabel users.
type UserModel struct {
	ID       string `gorm:"type:varchar(36);primaryKey"`
	Username string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (UserModel) TableName() string { return "users" }

// BookModel -> tabel books.
type BookModel struct {
	ID            string `gorm:"type:varchar(36);primaryKey"`
	Title         string `gorm:"type:varchar(255);not null"`
	Author        string `gorm:"type:varchar(255);not null"`
	PublishedDate time.Time
	Status        string `gorm:"type:varchar(20);index"`
	OwnerID       string `gorm:"type:varchar(36);index"`
	UserID        string `gorm:"type:varchar(36);index"`
}

func (BookModel) TableName() string { return "books" }

// BorrowedBookModel -> tabel borrowed_books.
type BorrowedBookModel struct {
	ID           string `gorm:"type:varchar(36);primaryKey"`
	BookID       string `gorm:"type:varchar(36);index"`
	UserID       string `gorm:"type:varchar(36);index"`
	BorrowedDate time.Time
	DueDate      time.Time
	ReturnDate   *time.Time
}

func (BorrowedBookModel) TableName() string { return "borrowed_books" }

// ---- Konversi model <-> domain -------------------

func toDomainUser(m *UserModel) *domain.User {
	return &domain.User{ID: m.ID, Username: m.Username, Password: m.Password}
}

func toDomainBook(m *BookModel) *domain.Book {
	return &domain.Book{
		ID:            m.ID,
		Title:         m.Title,
		Author:        m.Author,
		PublishedDate: m.PublishedDate,
		Status:        domain.BookStatus(m.Status),
		OwnerID:       m.OwnerID,
		UserID:        m.UserID,
	}
}

func toDomainBorrow(m *BorrowedBookModel) *domain.BorrowedBook {
	return &domain.BorrowedBook{
		ID:           m.ID,
		BookID:       m.BookID,
		UserID:       m.UserID,
		BorrowedDate: m.BorrowedDate,
		DueDate:      m.DueDate,
		ReturnDate:   m.ReturnDate,
	}
}

// NewDB membuka koneksi Postgres dan menjalankan AutoMigrate.
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&UserModel{}, &BookModel{}, &BorrowedBookModel{}); err != nil {
		return nil, err
	}
	return db, nil
}
