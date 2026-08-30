// Package jobs memasang job greet ke aplikasi: handler terdaftar di
// Queue, route POST /jobs/greet di Server httpx.
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sonix-framework/core/app"
	"github.com/sonix-framework/core/container"
	"github.com/sonix-framework/core/httpx"
	"github.com/sonix-framework/core/queue"
)

// Provider memasang demo job greet ke Queue dan Server.
type Provider struct{}

// Register menyediakan Greeter ke container dan menyatakan root-nya.
func (p *Provider) Register(c *container.Container) error {
	if err := c.Provide(NewGreeter); err != nil {
		return err
	}
	c.Root(func(*Greeter) {})
	return nil
}

// Boot mendaftarkan handler greet ke Queue dan route POST /jobs/greet
// ke Server. Route men-decode body, meneruskannya sebagai payload job,
// dan langsung 202 (pekerjaan selesai oleh worker).
func (p *Provider) Boot(ctx context.Context, a *app.Application) error {
	return a.Container().Invoke(func(g *Greeter, q *queue.Queue, srv *httpx.Server) error {
		if err := q.Register("greet", g.HandleGreet); err != nil {
			return fmt.Errorf("queue: daftar handler greet: %w", err)
		}
		return srv.HandleFunc("POST /jobs/greet", func(w http.ResponseWriter, r *http.Request) {
			var in greetIn
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "body JSON tidak valid")
				return
			}
			payload, err := json.Marshal(in)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			// Request ctx dipakai: buffer penuh → backpressure sampai
			// ada ruang atau klien pergi.
			if err := q.Dispatch(r.Context(), queue.Job{Name: "greet", Payload: payload}); err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "dispatch_failed", "gagal memasukkan antrean")
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})
	})
}
