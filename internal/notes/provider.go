package notes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sonix-framework/core/app"
	"github.com/sonix-framework/core/container"
	"github.com/sonix-framework/core/httpx"
)

// Provider memasang CRUD notes ke server HTTP aplikasi.
type Provider struct{}

// Register menyediakan Service dan API ke container.
func (p *Provider) Register(c *container.Container) error {
	if err := c.Provide(NewService); err != nil {
		return err
	}
	if err := c.Provide(NewAPI); err != nil {
		return err
	}
	c.Root(func(*Service, *API) {})
	return nil
}

// Boot mendaftarkan lima route CRUD ke Server httpx.
func (p *Provider) Boot(ctx context.Context, a *app.Application) error {
	return a.Container().Invoke(func(api *API, srv *httpx.Server) error {
		routes := []struct {
			pattern string
			h       http.HandlerFunc
		}{
			{"GET /notes", api.list},
			{"POST /notes", api.create},
			{"GET /notes/{id}", api.get},
			{"PUT /notes/{id}", api.update},
			{"DELETE /notes/{id}", api.remove},
		}
		for _, rt := range routes {
			if err := srv.HandleFunc(rt.pattern, rt.h); err != nil {
				return fmt.Errorf("httpx: daftar route %s: %w", rt.pattern, err)
			}
		}
		return nil
	})
}
