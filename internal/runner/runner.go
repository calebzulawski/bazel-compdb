package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bazel-compdb/internal/config"

	analysispb "bazel-compdb/internal/analysis_go_proto"
	"google.golang.org/protobuf/proto"
)

var sourceExtensions = map[string][]string{
	"CppCompile": {"c", "cc", "cpp", "cxx", "c++"},
	"ObjcCompile": {
		"c", "cc", "cpp", "cxx", "c++",
		"m", "mm",
	},
	"CudaCompile": {"cu", "cc", "cpp"},
}

func Run(ctx context.Context, cfg *config.Options) error {
	cmdArgs := assembleBazelArgs(cfg)
	bazelBinary := cfg.BazelBinary
	if bazelBinary == "" {
		bazelBinary = "bazel"
	}

	workspace, err := workspaceDir(ctx, bazelBinary)
	if err != nil {
		return fmt.Errorf("determine workspace: %w", err)
	}

	cmd := exec.CommandContext(ctx, bazelBinary, cmdArgs...)
	cmd.Dir = workspace
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bazel %s failed: %w\n%s", strings.Join(cmdArgs, " "), err, stderr.String())
	}

	var container analysispb.ActionGraphContainer
	if err := proto.Unmarshal(stdout.Bytes(), &container); err != nil {
		return fmt.Errorf("decode aquery proto: %w", err)
	}

	var commands []compileCommand

	for _, action := range container.GetActions() {
		args := action.GetArguments()
		if len(args) == 0 {
			continue
		}
		source := findSourceArg(action.GetMnemonic(), args)
		if source == "" {
			continue
		}
		output := findOutputArg(args)
		if output == "" {
			continue
		}
		commands = append(commands, compileCommand{
			Directory: workspace,
			Arguments: append([]string(nil), args...),
			File:      source,
			Output:    output,
		})
	}

	if err := writeCompileCommands(workspace, commands); err != nil {
		return err
	}

	return linkExternal(ctx, bazelBinary, workspace)
}

func assembleBazelArgs(cfg *config.Options) []string {
	args := []string{"aquery"}
	if len(cfg.BazelFlags) > 0 {
		args = append(args, cfg.BazelFlags...)
	}
	args = append(args, "--output=proto")

	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []string{"//..."}
	}
	wrapped := make([]string, len(targets))
	for i, tgt := range targets {
		wrapped[i] = fmt.Sprintf("mnemonic(\"(Cpp|Objc|Cuda)Compile\", %s)", tgt)
	}
	args = append(args, wrapped...)
	return args
}

func findSourceArg(mnemonic string, args []string) string {
	exts, ok := sourceExtensions[mnemonic]
	if !ok {
		return ""
	}
	for _, arg := range args {
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		lower := strings.ToLower(arg)
		for _, ext := range exts {
			if strings.HasSuffix(lower, "."+ext) {
				return arg
			}
		}
	}
	return ""
}

func findOutputArg(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "" {
			continue
		}
		switch {
		case arg == "-o" || arg == "--output":
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(arg, "-o") && len(arg) > 2:
			return arg[2:]
		case strings.HasPrefix(arg, "--output="):
			val := arg[len("--output="):]
			if val != "" {
				return val
			}
		case strings.HasPrefix(arg, "/Fo"):
			val := arg[3:]
			if val != "" {
				return val
			}
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(arg, "-Fo"):
			val := arg[3:]
			if val != "" {
				return val
			}
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

func workspaceDir(ctx context.Context, bazelBinary string) (string, error) {
	cmd := exec.CommandContext(ctx, bazelBinary, "info", "workspace")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("bazel info workspace failed: %w\n%s", err, stderr.String())
	}
	workspace := strings.TrimSpace(stdout.String())
	if workspace == "" {
		return "", fmt.Errorf("bazel info workspace returned empty path")
	}
	return workspace, nil
}

func linkExternal(ctx context.Context, bazelBinary, workspace string) error {
	cmd := exec.CommandContext(ctx, bazelBinary, "info", "execution_root")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bazel info execution_root failed: %w\n%s", err, stderr.String())
	}
	execRoot := strings.TrimSpace(stdout.String())
	if execRoot == "" {
		return fmt.Errorf("bazel info execution_root returned empty path")
	}

	linkPath := filepath.Join(workspace, "external")
	target := filepath.Join(execRoot, "external")

	if _, err := os.Lstat(linkPath); err == nil {
		return nil
	}

	if err := os.Symlink(target, linkPath); err != nil {
		return fmt.Errorf("create symlink %s -> %s: %w", linkPath, target, err)
	}
	return nil
}
