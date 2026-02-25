package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type compileCommand struct {
	Directory string   `json:"directory"`
	Arguments []string `json:"arguments"`
	File      string   `json:"file"`
	Output    string   `json:"output"`
}

func writeCompileCommands(workspace, outputPath string, commands []compileCommand) error {
	outputPath = filepath.Clean(outputPath)
	path := outputPath
	if filepath.IsAbs(outputPath) {
		path = outputPath
	} else {
		path = filepath.Join(workspace, outputPath)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal compile commands: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
