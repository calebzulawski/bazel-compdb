package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// splitBazelArgs resolves Bazel command-line args into Bazel flags and target
// patterns by asking Bazel to parse the command via BEP.
func splitBazelArgs(args []string, bazelBinary, cwd string) ([]string, []string, error) {
	if len(args) == 0 {
		return nil, nil, nil
	}
	if bazelBinary == "" {
		bazelBinary = "bazel"
	}

	eventPath, err := runBazelParseBEP(args, bazelBinary, cwd)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(eventPath)

	eventFile, err := os.Open(eventPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open build event file %s: %w", eventPath, err)
	}
	defer eventFile.Close()

	targets, err := parseBEPTargets(eventFile)
	if err != nil {
		return nil, nil, err
	}
	flags := filterTargetsFromArgs(args, targets)
	return flags, targets, nil
}

func parseBEPTargets(r io.Reader) ([]string, error) {
	var targets []string
	seenTargets := map[string]struct{}{}

	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			line = bytes.TrimSpace(line)
			if len(line) > 0 {
				var ev bepEvent
				if unmarshalErr := json.Unmarshal(line, &ev); unmarshalErr != nil {
					return nil, fmt.Errorf("parse build event: %w", unmarshalErr)
				}
				collectBEPTargets(ev, seenTargets, &targets)
			}
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read build event stream: %w", err)
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var ev bepEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, fmt.Errorf("parse build event: %w", err)
		}
		collectBEPTargets(ev, seenTargets, &targets)
	}

	return targets, nil
}

func collectBEPTargets(ev bepEvent, seenTargets map[string]struct{}, targets *[]string) {
	if ev.ID.Pattern == nil {
		return
	}
	for _, pattern := range ev.ID.Pattern.Pattern {
		if _, exists := seenTargets[pattern]; exists {
			continue
		}
		seenTargets[pattern] = struct{}{}
		*targets = append(*targets, pattern)
	}
}

type bepEvent struct {
	ID struct {
		Pattern *struct {
			Pattern []string `json:"pattern"`
		} `json:"pattern"`
	} `json:"id"`
}

func runBazelParseBEP(args []string, bazelBinary, cwd string) (string, error) {
	eventFile, err := os.CreateTemp("", "bazel-compdb-bep-*.json")
	if err != nil {
		return "", fmt.Errorf("create temporary build event file: %w", err)
	}
	eventPath := eventFile.Name()
	if err := eventFile.Close(); err != nil {
		os.Remove(eventPath)
		return "", fmt.Errorf("close temporary build event file: %w", err)
	}

	cmdArgs := []string{
		"--ignore_all_rc_files",
		"build",
		"--nobuild",
		"--build_event_json_file=" + eventPath,
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(bazelBinary, cmdArgs...)
	cmd.Dir = cwd
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		os.Remove(eventPath)
		return "", fmt.Errorf("parse Bazel args with BEP failed: %w\n%s", err, stderr.String())
	}
	return eventPath, nil
}

func filterTargetsFromArgs(args []string, targets []string) []string {
	if len(args) == 0 {
		return nil
	}
	targetSet := make(map[string]struct{}, len(targets))
	for _, tgt := range targets {
		targetSet[tgt] = struct{}{}
	}

	flags := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--" {
			continue
		}
		if _, isTarget := targetSet[arg]; isTarget {
			continue
		}
		flags = append(flags, arg)
	}
	return flags
}
