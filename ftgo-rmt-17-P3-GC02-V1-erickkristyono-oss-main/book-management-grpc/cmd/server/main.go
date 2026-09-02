
// Menjalankan: server gRPC, REST gateway (grpc-gateway) + Swagger, dan job scheduler.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/book-management-grpc/internal/config"
	grpcdelivery "github.com/yourusername/book-management-grpc/internal/delivery/grpc"
	"github.com/yourusername/book-management-grpc/internal/repository/memory"
	"github.com/yourusername/book-management-grpc/internal/repository/postgres"
	"github.com/yourusername/book-management-grpc/internal/scheduler"
	"github.com/yourusername/book-management-grpc/internal/security"
	"github.com/yourusername/book-management-grpc/internal/token"
	"github.com/yourusername/book-management-grpc/internal/usecase"
)

func main() {
	cfg := config.Load()
	log.Printf("konfigurasi: driver=%s grpc=:%s http=:%s", cfg.DBDriver, cfg.GRPCPort, cfg.HTTPPort)

	// --- Pilih implementasi repository (in-memory / postgres) ---------------
	var (
		userRepo   usecase.UserRepository
		bookRepo   usecase.BookRepository
		borrowRepo usecase.BorrowRepository
	)
	switch cfg.DBDriver {
	case "postgres":
		db, err := postgres.NewDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("gagal konek postgres: %v", err)
		}
		userRepo = postgres.NewUserRepository(db)
		bookRepo = postgres.NewBookRepository(db)
		borrowRepo = postgres.NewBorrowRepository(db)
		log.Println("menggunakan repository: postgres")
	default:
		userRepo = memory.NewUserRepository()
		bookRepo = memory.NewBookRepository()
		borrowRepo = memory.NewBorrowRepository()
		log.Println("menggunakan repository: in-memory")
	}

	// --- Dependency injection (Clean Architecture) --------------------------
	hasher := security.NewHasher()
	tokens := token.NewManager(cfg.JWTSecret, cfg.JWTTTL)

	authUC := usecase.NewAuthUsecase(userRepo, hasher, tokens)
	bookUC := usecase.NewBookUsecase(bookRepo, borrowRepo)

	authH := grpcdelivery.NewAuthHandler(authUC, tokens)
	bookH := grpcdelivery.NewBookHandler(bookUC)
	interceptor := grpcdelivery.NewAuthInterceptor(tokens)

	grpcServer := grpcdelivery.NewGRPCServer(authH, bookH, interceptor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Jalankan server gRPC ----------------------------------------------
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("gagal listen gRPC: %v", err)
	}
	go func() {
		log.Printf("gRPC server berjalan di :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server berhenti: %v", err)
		}
	}()

	// --- Jalankan REST gateway + Swagger -----------------------------------
	gwHandler, err := grpcdelivery.NewGatewayHandler(ctx, "localhost:"+cfg.GRPCPort, "docs/swagger")
	if err != nil {
		log.Fatalf("gagal membuat gateway: %v", err)
	}
	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: gwHandler}
	go func() {
		log.Printf("HTTP gateway + Swagger di :%s (Swagger UI: http://localhost:%s/swagger/)", cfg.HTTPPort, cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server berhenti: %v", err)
		}
	}()

	// --- Jalankan job scheduler (update buku overdue) ----------------------
	job := func(ctx context.Context) (int, error) {
		return bookUC.MarkOverdueBooks(ctx)
	}
	sched := scheduler.New("overdue-updater", cfg.SchedInterval, job, scheduler.WithRunOnStart(true))
	sched.Start(ctx)
	log.Printf("job scheduler aktif (interval %s)", cfg.SchedInterval)

	// --- Graceful shutdown --------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("menerima sinyal shutdown, membersihkan...")

	cancel() // hentikan scheduler & gateway context

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
	sched.Stop()
	log.Println("shutdown selesai.")
}
