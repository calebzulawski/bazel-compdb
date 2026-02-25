package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazel-compdb/internal/config"
	"bazel-compdb/internal/runner"
)

// moduleVersion is set via x_defs on the Bazel binary target.
var moduleVersion = ""

// buildVersion is injected at link time for release binaries.
var buildVersion = ""

func reportedVersion() string {
	if buildVersion != "" {
		return buildVersion
	}
	return moduleVersion
}

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
		if errors.Is(err, config.ErrHelp) {
			os.Exit(0)
		}
		if errors.Is(err, config.ErrVersion) {
			fmt.Println(reportedVersion())
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := runner.Run(ctx, cfg); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
