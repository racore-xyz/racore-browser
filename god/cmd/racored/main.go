package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/racore/god/internal/config"
	"github.com/racore/god/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()
	cfg.DataDir = config.DataDir()

	srv := server.New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("start: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	srv.Stop()
}
