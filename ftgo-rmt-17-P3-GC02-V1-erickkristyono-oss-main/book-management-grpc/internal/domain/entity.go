// Package domain berisi entitas inti (enterprise business rules) dari aplikasi.
// Layer ini TIDAK bergantung pada layer lain (paling dalam pada Clean Architecture).
package domain

import "time"

// BookStatus merepresentasikan status sebuah buku.
type BookStatus string

const (
	// StatusAvailable: buku tersedia untuk dipinjam.
	StatusAvailable BookStatus = "Available"
	// StatusBorrowed: buku sedang dipinjam dan masih dalam tenggat waktu.
	StatusBorrowed BookStatus = "Borrowed"
	// StatusOverdue: buku dipinjam namun melewati batas waktu pengembalian.
	// Status ini di-set otomatis oleh job scheduler.
	StatusOverdue BookStatus = "Overdue"
)

// User merepresentasikan pengguna yang terdaftar.
type User struct {
	ID       string // UUID
	Username string
	Password string // hash password (bukan plaintext)
}

// Book merepresentasikan sebuah buku pada sistem.
//
// Catatan desain:
//   - OwnerID  : pengguna yang menambahkan/ memiliki buku (dipakai untuk otorisasi update/delete).
//   - UserID   : pengguna yang sedang meminjam buku (sesuai skema soal; kosong jika Available).
type Book struct {
	ID            string
	Title         string
	Author        string
	PublishedDate time.Time
	Status        BookStatus
	OwnerID       string // pemilik/penambah buku
	UserID        string // peminjam saat ini (kosong bila tidak dipinjam)
}

// BorrowedBook merepresentasikan satu catatan peminjaman.
type BorrowedBook struct {
	ID           string
	BookID       string
	UserID       string
	BorrowedDate time.Time
	DueDate      time.Time  // batas waktu pengembalian (untuk job scheduler)
	ReturnDate   *time.Time // nil bila belum dikembalikan
}

// IsReturned mengembalikan true jika peminjaman sudah dikembalikan.
func (b *BorrowedBook) IsReturned() bool { return b.ReturnDate != nil }

// IsOverdue mengembalikan true jika peminjaman belum dikembalikan dan
// sudah melewati DueDate relatif terhadap waktu now.
func (b *BorrowedBook) IsOverdue(now time.Time) bool {
	return !b.IsReturned() && now.After(b.DueDate)
}
