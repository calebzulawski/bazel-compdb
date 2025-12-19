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

func writeCompileCommands(workspace string, commands []compileCommand) error {
	path := filepath.Join(workspace, "compile_commands.json")
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal compile_commands.json: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write compile_commands.json: %w", err)
	}
	return nil
}
