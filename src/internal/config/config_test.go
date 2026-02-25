package config

import (
	"errors"
	"reflect"
	"testing"
)

func TestSplitArgs_WithSeparator(t *testing.T) {
	t.Parallel()

	args := []string{"--bazel", "/usr/bin/bazel", "--", "-c", "dbg", "//..."}
	inv := parseInvocation(args)

	if !reflect.DeepEqual(inv.toolArgs, []string{"--bazel", "/usr/bin/bazel"}) {
		t.Fatalf("toolArgs=%v", inv.toolArgs)
	}
	if !reflect.DeepEqual(inv.bazelArgs, []string{"-c", "dbg", "//..."}) {
		t.Fatalf("bazelArgs=%v", inv.bazelArgs)
	}
}

func TestSplitArgs_WithoutSeparatorTreatsAllAsBazelArgs(t *testing.T) {
	t.Parallel()

	args := []string{"-c", "dbg", "--remote_download_outputs=minimal", "//foo:bar"}
	inv := parseInvocation(args)

	if inv.toolArgs != nil {
		t.Fatalf("toolArgs=%v, want nil", inv.toolArgs)
	}
	if !reflect.DeepEqual(inv.bazelArgs, args) {
		t.Fatalf("bazelArgs=%v, want=%v", inv.bazelArgs, args)
	}
}

func TestSplitArgs_WithoutSeparatorHelpIsProgramArg(t *testing.T) {
	t.Parallel()

	for _, helpArg := range []string{"--help", "-h"} {
		inv := parseInvocation([]string{helpArg})
		if !reflect.DeepEqual(inv.toolArgs, []string{helpArg}) {
			t.Fatalf("toolArgs=%v", inv.toolArgs)
		}
		if inv.bazelArgs != nil {
			t.Fatalf("bazelArgs=%v, want nil", inv.bazelArgs)
		}
	}
}

func TestSplitArgs_WithoutSeparatorVersionIsProgramArg(t *testing.T) {
	t.Parallel()

	for _, versionArg := range []string{"--version", "-v"} {
		inv := parseInvocation([]string{versionArg})
		if !reflect.DeepEqual(inv.toolArgs, []string{versionArg}) {
			t.Fatalf("toolArgs=%v", inv.toolArgs)
		}
		if inv.bazelArgs != nil {
			t.Fatalf("bazelArgs=%v, want nil", inv.bazelArgs)
		}
	}
}

func TestParseToolFlags_ParsesShortOutputAlias(t *testing.T) {
	t.Parallel()

	cli, err := parseToolFlags([]string{"--bazel=/usr/bin/bazel", "-o", "out/compile_commands.json"})
	if err != nil {
		t.Fatalf("parseToolFlags returned error: %v", err)
	}
	if cli.BazelBinary != "/usr/bin/bazel" {
		t.Fatalf("BazelBinary=%q", cli.BazelBinary)
	}
	if cli.OutputPath != "out/compile_commands.json" {
		t.Fatalf("OutputPath=%q", cli.OutputPath)
	}
}

func TestParseToolFlags_VersionReturnsErrVersion(t *testing.T) {
	t.Parallel()

	_, err := parseToolFlags([]string{"-v"})
	if !errors.Is(err, ErrVersion) {
		t.Fatalf("err=%v, want ErrVersion", err)
	}
}
