package grpcdelivery

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yourusername/book-management-grpc/internal/domain"
	"github.com/yourusername/book-management-grpc/internal/usecase"
	bookv1 "github.com/yourusername/book-management-grpc/pkg/genproto/book/v1"
)

// BookHandler mengimplementasikan bookv1.BookServiceServer.
type BookHandler struct {
	bookv1.UnimplementedBookServiceServer
	books *usecase.BookUsecase
}

// NewBookHandler membuat BookHandler.
func NewBookHandler(books *usecase.BookUsecase) *BookHandler {
	return &BookHandler{books: books}
}

func toProtoBook(b *domain.Book) *bookv1.Book {
	var pd *timestamppb.Timestamp
	if !b.PublishedDate.IsZero() {
		pd = timestamppb.New(b.PublishedDate)
	}
	return &bookv1.Book{
		Id:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		PublishedDate: pd,
		Status:        string(b.Status),
		OwnerId:       b.OwnerID,
		UserId:        b.UserID,
	}
}

func mustUser(ctx context.Context) (string, error) {
	uid, ok := UserIDFromContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user tidak terautentikasi")
	}
	return uid, nil
}

// CreateBook menambah buku milik user yang login.
func (h *BookHandler) CreateBook(ctx context.Context, req *bookv1.CreateBookRequest) (*bookv1.Book, error) {
	uid, err := mustUser(ctx)
	if err != nil {
		return nil, err
	}
	in := usecase.CreateBookInput{
		Title:   req.GetTitle(),
		Author:  req.GetAuthor(),
		OwnerID: uid,
	}
	if ts := req.GetPublishedDate(); ts != nil {
		in.PublishedDate = ts.AsTime()
	}
	b, err := h.books.CreateBook(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoBook(b), nil
}

// GetBook mengambil satu buku.
func (h *BookHandler) GetBook(ctx context.Context, req *bookv1.GetBookRequest) (*bookv1.Book, error) {
	b, err := h.books.GetBook(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoBook(b), nil
}

// ListBooks mengambil daftar buku (opsional difilter status).
func (h *BookHandler) ListBooks(ctx context.Context, req *bookv1.ListBooksRequest) (*bookv1.ListBooksResponse, error) {
	f := usecase.BookFilter{Status: domain.BookStatus(req.GetStatus())}
	books, err := h.books.ListBooks(ctx, f)
	if err != nil {
		return nil, toGRPCError(err)
	}
	resp := &bookv1.ListBooksResponse{Books: make([]*bookv1.Book, 0, len(books))}
	for _, b := range books {
		resp.Books = append(resp.Books, toProtoBook(b))
	}
	return resp, nil
}

// UpdateBook memperbarui buku (hanya pemilik).
func (h *BookHandler) UpdateBook(ctx context.Context, req *bookv1.UpdateBookRequest) (*bookv1.Book, error) {
	uid, err := mustUser(ctx)
	if err != nil {
		return nil, err
	}
	in := usecase.UpdateBookInput{
		ID:      req.GetId(),
		Title:   req.GetTitle(),
		Author:  req.GetAuthor(),
		ActorID: uid,
	}
	if ts := req.GetPublishedDate(); ts != nil {
		in.PublishedDate = ts.AsTime()
	}
	b, err := h.books.UpdateBook(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoBook(b), nil
}

// DeleteBook menghapus buku (hanya pemilik).
func (h *BookHandler) DeleteBook(ctx context.Context, req *bookv1.DeleteBookRequest) (*bookv1.DeleteBookResponse, error) {
	uid, err := mustUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.books.DeleteBook(ctx, req.GetId(), uid); err != nil {
		return nil, toGRPCError(err)
	}
	return &bookv1.DeleteBookResponse{Success: true}, nil
}

// BorrowBook meminjam buku.
func (h *BookHandler) BorrowBook(ctx context.Context, req *bookv1.BorrowBookRequest) (*bookv1.BorrowResponse, error) {
	uid, err := mustUser(ctx)
	if err != nil {
		return nil, err
	}
	dur := time.Duration(req.GetDurationHours()) * time.Hour
	rec, err := h.books.BorrowBook(ctx, req.GetBookId(), uid, dur)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &bookv1.BorrowResponse{
		BorrowId: rec.ID,
		BookId:   rec.BookID,
		DueDate:  timestamppb.New(rec.DueDate),
	}, nil
}

// ReturnBook mengembalikan buku.
func (h *BookHandler) ReturnBook(ctx context.Context, req *bookv1.ReturnBookRequest) (*bookv1.ReturnResponse, error) {
	uid, err := mustUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.books.ReturnBook(ctx, req.GetBookId(), uid); err != nil {
		return nil, toGRPCError(err)
	}
	return &bookv1.ReturnResponse{Success: true}, nil
}
