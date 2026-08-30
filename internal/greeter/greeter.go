// Package greeter adalah contoh lengkap satu kemampuan aplikasi:
// service + provider + runner.
package greeter

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/sonix-framework/core/app"
	"github.com/sonix-framework/core/config"
	"github.com/sonix-framework/core/container"
)

// Service adalah contoh service aplikasi.
// Dependen diterima lewat constructor — service tidak tahu container ada.
type Service struct {
	log  *slog.Logger
	name string
}

// NewService dipakai container sebagai constructor.
func NewService(cfg *config.Repository, log *slog.Logger) (*Service, error) {
	name := cfg.String("app.name", "sonix")
	if name == "" {
		return nil, errors.New("greeter: app.name kosong")
	}
	return &Service{log: log, name: name}, nil
}

// Greet contoh method bisnis.
func (s *Service) Greet() string { return "Halo dari " + s.name }

// Provider mendaftarkan Service ke container dan menjalankan heartbeat saat boot.
type Provider struct{}

// Register hanya deklaratif: daftarkan constructor + root validasi.
func (p *Provider) Register(c *container.Container) error {
	if err := c.Provide(NewService); err != nil {
		return err
	}
	// Root menyatakan: graph Service wajib sehat sebelum aplikasi jalan.
	c.Root(func(*Service) {})
	return nil
}

// Boot mengambil instance Service lewat container dan mendaftarkan Runner.
func (p *Provider) Boot(ctx context.Context, a *app.Application) error {
	var svc *Service
	if err := a.Container().Invoke(func(s *Service) { svc = s }); err != nil {
		return err
	}
	interval := a.Config().Duration("heartbeat.interval", 2*time.Second)
	a.RegisterRunner(&Heartbeat{svc: svc, log: a.Logger(), interval: interval})
	return nil
}

// Heartbeat membuktikan lifecycle: jalan sampai signal, mati dengan rapi.
type Heartbeat struct {
	svc      *Service
	log      *slog.Logger
	interval time.Duration
}

// Run berhenti ketika context dibatalkan — cancel bukan error.
func (h *Heartbeat) Run(ctx context.Context) error {
	t := time.NewTicker(h.interval)
	defer t.Stop()
	h.log.Info("heartbeat aktif", "interval", h.interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			h.log.Info("heartbeat", "pesan", h.svc.Greet())
		}
	}
}

// Shutdown adalah pamit terakhir sebelum proses berhenti.
func (h *Heartbeat) Shutdown(ctx context.Context) error {
	h.log.Info("heartbeat dimatikan", "pesan", h.svc.Greet())
	return nil
}
