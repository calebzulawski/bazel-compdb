package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

const defaultConfigName = ".bazel-compdb"

var (
	ErrHelp    = errors.New("help requested")
	ErrVersion = errors.New("version requested")
)

type Options struct {
	BazelBinary string   `yaml:"bazel"`
	BazelFlags  []string `yaml:"bazel_flags"`
	Targets     []string `yaml:"targets"`
	OutputPath  string   `yaml:"output"`
}

type cliOptions struct {
	BazelBinary string `name:"bazel" help:"Path to the bazel binary."`
	Help        bool   `name:"help" short:"h" help:"Show this help."`
	OutputPath  string `name:"output" short:"o" help:"Path to write compile_commands.json."`
	Version     bool   `name:"version" short:"v" help:"Show version."`
}

type invocation struct {
	toolArgs  []string
	bazelArgs []string
}

func Parse() (*Options, error) {
	inv := parseInvocation(os.Args[1:])
	cliOpts, err := parseToolFlags(inv.toolArgs)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to determine current directory: %w", err)
	}

	fileCfg, err := loadConfigs(cwd)
	if err != nil {
		return nil, err
	}

	parseBazelBinary := cliOpts.BazelBinary
	if parseBazelBinary == "" {
		parseBazelBinary = fileCfg.BazelBinary
	}

	bazelFlagOpts, bazelTargetOpts, err := splitBazelArgs(inv.bazelArgs, parseBazelBinary, cwd)
	if err != nil {
		return nil, err
	}

	overlay := Options{
		BazelBinary: cliOpts.BazelBinary,
		OutputPath:  cliOpts.OutputPath,
	}
	overlay.BazelFlags = bazelFlagOpts
	overlay.Targets = bazelTargetOpts

	merged := mergeOptions(fileCfg, overlay)
	applyDefaults(&merged)

	return &merged, nil
}

func parseToolFlags(args []string) (cliOptions, error) {
	var cli cliOptions
	parser, err := kong.New(
		&cli,
		kong.Name("bazel-compdb"),
		kong.NoDefaultHelp(),
		kong.Help(helpPrinter),
	)
	if err != nil {
		return cliOptions{}, err
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		return cliOptions{}, err
	}
	if cli.Help {
		if err := ctx.PrintUsage(false); err != nil {
			return cliOptions{}, err
		}
		return cliOptions{}, ErrHelp
	}
	if cli.Version {
		return cliOptions{}, ErrVersion
	}
	return cli, nil
}

func helpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	fmt.Fprintln(ctx.Stdout, "Usage:")
	fmt.Fprintln(ctx.Stdout, "  bazel-compdb [arguments] -- [bazel arguments]")
	fmt.Fprintln(ctx.Stdout, "  bazel-compdb [bazel arguments]")
	fmt.Fprintln(ctx.Stdout)

	options.NoAppSummary = true
	return kong.DefaultHelpPrinter(options, ctx)
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
	if override.OutputPath != "" {
		result.OutputPath = override.OutputPath
	}
	return result
}

func parseInvocation(args []string) invocation {
	for idx, arg := range args {
		if arg == "--" {
			return invocation{
				toolArgs:  args[:idx],
				bazelArgs: args[idx+1:],
			}
		}
	}
	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help", "-v", "--version":
			return invocation{toolArgs: args}
		}
	}
	if len(args) == 0 {
		return invocation{toolArgs: args}
	}

	// Without an explicit separator, treat all CLI args as Bazel args.
	return invocation{bazelArgs: args}
}

func applyDefaults(opts *Options) {
	if len(opts.Targets) == 0 {
		opts.Targets = []string{"//..."}
	}
	if opts.OutputPath == "" {
		opts.OutputPath = "compile_commands.json"
	}
}
