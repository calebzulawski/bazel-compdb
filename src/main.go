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

	if wd := os.Getenv("BUILD_WORKING_DIRECTORY"); wd != "" {
		if err := os.Chdir(wd); err != nil {
			log.Fatalf("application error: change directory to BUILD_WORKING_DIRECTORY (%s): %v", wd, err)
		}
	}

	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := runner.Run(ctx, cfg); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
