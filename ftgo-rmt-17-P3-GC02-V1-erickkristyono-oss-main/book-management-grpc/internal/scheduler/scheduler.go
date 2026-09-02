// Package scheduler menjalankan pekerjaan (job) secara berkala.
// Implementasi memakai time.Ticker dari standard library sehingga ringan,
// mudah diuji, dan tidak butuh dependency eksternal.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job adalah fungsi pekerjaan yang dijalankan berkala.
// Mengembalikan jumlah item yang diproses dan error (bila ada).
type Job func(ctx context.Context) (int, error)

// Scheduler menjalankan sebuah Job pada interval tertentu.
type Scheduler struct {
	name     string
	interval time.Duration
	job      Job
	logger   *log.Logger
	runNow   bool

	wg sync.WaitGroup
}

// Option mengonfigurasi Scheduler.
type Option func(*Scheduler)

// WithLogger mengganti logger.
func WithLogger(l *log.Logger) Option { return func(s *Scheduler) { s.logger = l } }

// WithRunOnStart menentukan apakah job langsung dijalankan sekali saat Start.
func WithRunOnStart(v bool) Option { return func(s *Scheduler) { s.runNow = v } }

// New membuat Scheduler baru.
func New(name string, interval time.Duration, job Job, opts ...Option) *Scheduler {
	s := &Scheduler{
		name:     name,
		interval: interval,
		job:      job,
		logger:   log.Default(),
		runNow:   true,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Start menjalankan loop scheduler di goroutine terpisah.
// Loop berhenti ketika ctx dibatalkan. Gunakan Stop() untuk menunggu selesai.
func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if s.runNow {
			s.runOnce(ctx)
		}
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.logger.Printf("[scheduler:%s] berhenti", s.name)
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

// runOnce mengeksekusi job sekali sambil menangani panic & error.
func (s *Scheduler) runOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Printf("[scheduler:%s] panic dipulihkan: %v", s.name, r)
		}
	}()
	n, err := s.job(ctx)
	if err != nil {
		s.logger.Printf("[scheduler:%s] error: %v", s.name, err)
		return
	}
	if n > 0 {
		s.logger.Printf("[scheduler:%s] selesai, %d item diperbarui", s.name, n)
	}
}

// Stop menunggu goroutine scheduler benar-benar berhenti (setelah ctx dibatalkan).
func (s *Scheduler) Stop() { s.wg.Wait() }
