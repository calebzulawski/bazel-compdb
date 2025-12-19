package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazel-compdb/internal/config"
	"bazel-compdb/internal/runner"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := runner.Run(ctx, cfg); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
