// Package bootstrap merangkai aplikasi Sonix: ini composition root-nya.
package bootstrap

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/sonix-framework/core/app"
	"github.com/sonix-framework/core/config"
	"github.com/sonix-framework/core/httpx"
	"github.com/sonix-framework/core/logging"

	"github.com/sonix-framework/sonix/internal/greeter"
	"github.com/sonix-framework/sonix/internal/notes"
)

// CreateApplication menghasilkan Application yang sudah siap Boot & Run.
func CreateApplication() (*app.Application, error) {
	// 1. Config: precedence defaults -> file -> environment.
	cfg := config.New()
	err := cfg.Defaults(map[string]any{
		"app.name":              "sonix",
		"logging.level":         "info",
		"logging.format":        "text",
		"heartbeat.interval":    "2s",
		"http.port":             8080,
		"http.read_timeout":     "10s",
		"http.write_timeout":    "10s",
		"http.idle_timeout":     "120s",
		"http.shutdown_timeout": "10s",
	})
	if err != nil {
		return nil, fmt.Errorf("bootstrap: defaults: %w", err)
	}
	if _, err := os.Stat("config/app.yaml"); err == nil {
		if err := cfg.File("config/app.yaml"); err != nil {
			return nil, fmt.Errorf("bootstrap: file config: %w", err)
		}
	}
	if err := cfg.Environment("SONIX_"); err != nil {
		return nil, fmt.Errorf("bootstrap: env config: %w", err)
	}

	// 2. Logging dibaca dari config.
	logger := logging.New(logging.Options{
		Level:  cfg.String("logging.level", "info"),
		Format: cfg.String("logging.format", "text"),
	})

	// 3. Kernel + dependensi inti untuk container.
	application := app.New(app.WithConfig(cfg), app.WithLogger(logger))
	c := application.Container()
	if err := c.Provide(func() *config.Repository { return cfg }); err != nil {
		return nil, fmt.Errorf("bootstrap: provide config: %w", err)
	}
	if err := c.Provide(func() *slog.Logger { return logger }); err != nil {
		return nil, fmt.Errorf("bootstrap: provide logger: %w", err)
	}

	if err := application.Register(&greeter.Provider{}); err != nil {
		return nil, fmt.Errorf("bootstrap: register provider %T: %w", &greeter.Provider{}, err)
	}
	if err := application.Register(&httpx.Provider{}); err != nil {
		return nil, fmt.Errorf("bootstrap: register provider %T: %w", &httpx.Provider{}, err)
	}
	if err := application.Register(&notes.Provider{}); err != nil {
		return nil, fmt.Errorf("bootstrap: register provider %T: %w", &notes.Provider{}, err)
	}

	return application, nil
}
