package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigName = ".bazel-compdb"

type Options struct {
	BazelBinary string   `yaml:"bazel"`
	BazelFlags  []string `yaml:"bazel_flags"`
	Targets     []string `yaml:"targets"`
}

func Parse() (*Options, error) {
	args := os.Args[1:]
	programArgs, bazelArgs := splitArgs(args)
	flagSet := flag.NewFlagSet("bazel-compdb", flag.ContinueOnError)

	bazelBinary := flagSet.String("bazel", "", "Path to the bazel binary")

	if err := flagSet.Parse(programArgs); err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to determine current directory: %w", err)
	}

	var merged Options

	fileCfg, err := loadConfigs(cwd)
	if err != nil {
		return nil, err
	}
	merged = mergeOptions(merged, fileCfg)

	merged = mergeOptions(merged, Options{
		BazelBinary: *bazelBinary,
	})
	bazelFlagOpts, bazelTargetOpts := splitBazelArgs(bazelArgs)
	merged = mergeOptions(merged, Options{
		BazelFlags: bazelFlagOpts,
		Targets:    bazelTargetOpts,
	})

	if len(merged.Targets) == 0 {
		merged.Targets = []string{"//..."}
	}

	return &merged, nil
}

func loadConfigs(workspace string) (Options, error) {
	var configPaths []string
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths, filepath.Join(home, defaultConfigName))
	}
	configPaths = append(configPaths, filepath.Join(workspace, defaultConfigName))

	var opts Options
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var fc Options
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return Options{}, fmt.Errorf("parse %s: %w", path, err)
		}
		opts = mergeOptions(opts, fc)
	}
	return opts, nil
}

func mergeOptions(base Options, override Options) Options {
	result := base
	if override.BazelBinary != "" {
		result.BazelBinary = override.BazelBinary
	}
	if override.BazelFlags != nil {
		result.BazelFlags = append([]string(nil), override.BazelFlags...)
	}
	if override.Targets != nil {
		result.Targets = append([]string(nil), override.Targets...)
	}
	return result
}

func splitBazelArgs(args []string) ([]string, []string) {
	if len(args) == 0 {
		return nil, nil
	}
	var flags []string
	var targets []string
	before, after := splitArgs(args)

	for _, arg := range before {
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
		} else {
			targets = append(targets, arg)
		}
	}

	if len(after) > 0 {
		joined := strings.TrimSpace(strings.Join(after, " "))
		if joined != "" {
			targets = append(targets, joined)
		}
	}

	return flags, targets
}

func splitArgs(args []string) ([]string, []string) {
	for idx, arg := range args {
		if arg == "--" {
			return args[:idx], args[idx+1:]
		}
	}
	return args, nil
}
