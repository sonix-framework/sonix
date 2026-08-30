package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sonix-framework/sonix/bootstrap"
)

func main() {
	// Signal context: Ctrl+C (SIGINT) atau SIGTERM membatalkan ctx.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := bootstrap.CreateApplication()
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap gagal:", err)
		os.Exit(1)
	}

	if err := application.Boot(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "boot gagal:", err)
		os.Exit(1)
	}

	// Run memblokir sampai signal; shutdown berlangsung di dalamnya.
	if err := application.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "aplikasi berhenti dengan error:", err)
		os.Exit(1)
	}
}
