// Package jobs adalah demo job queue: bukti M3 bahwa handler HTTP bisa
// memasukkan job ke antrean (Dispatch) dan worker core/queue
// memprosesnya dengan retry dan drain rapi.
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
)

// greetIn adalah payload job greet.
type greetIn struct {
	Nama  string `json:"nama"`
	Gagal bool   `json:"gagal"` // paksa gagal: demo retry
}

// Greeter adalah handler job greet; logger diterima lewat constructor —
// tanpa global logger (aturan AGENTS.md).
type Greeter struct {
	log *slog.Logger
}

// NewGreeter dipakai container sebagai constructor.
func NewGreeter(log *slog.Logger) *Greeter {
	return &Greeter{log: log}
}

// HandleGreet memproses payload greet; in.Gagal sengaja mengembalikan
// error untuk mendemokan retry worker.
func (g *Greeter) HandleGreet(ctx context.Context, payload []byte) error {
	var in greetIn
	if err := json.Unmarshal(payload, &in); err != nil {
		return errors.New("jobs: payload greet rusak")
	}
	if in.Gagal {
		return errors.New("jobs: greet gagal (simulasi)")
	}
	g.log.Info("halo dari queue", "nama", in.Nama)
	return nil
}
