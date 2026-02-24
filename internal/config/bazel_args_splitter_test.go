package config

import (
	"strings"
	"testing"
)

func TestParseBEPTargetsAndFilterFlags(t *testing.T) {
	bep := strings.Join([]string{
		`{"id":{"pattern":{"pattern":["//...","//foo:bar"]}}}`,
		`{"id":{"pattern":{"pattern":["//foo:bar",":local"]}}}`,
		"",
	}, "\n")

	targets, err := parseBEPTargets(strings.NewReader(bep))
	if err != nil {
		t.Fatalf("parseBEPTargets returned error: %v", err)
	}

	args := []string{"-c", "dbg", "--", "//...", "//foo:bar", ":local", "--copt=-Wall"}
	flags := filterTargetsFromArgs(args, targets)

	wantFlags := []string{"-c", "dbg", "--copt=-Wall"}
	if len(flags) != len(wantFlags) {
		t.Fatalf("flags length=%d, want=%d (%v)", len(flags), len(wantFlags), flags)
	}
	for i := range wantFlags {
		if flags[i] != wantFlags[i] {
			t.Fatalf("flags[%d]=%q, want=%q", i, flags[i], wantFlags[i])
		}
	}

	wantTargets := []string{"//...", "//foo:bar", ":local"}
	if len(targets) != len(wantTargets) {
		t.Fatalf("targets length=%d, want=%d (%v)", len(targets), len(wantTargets), targets)
	}
	for i := range wantTargets {
		if targets[i] != wantTargets[i] {
			t.Fatalf("targets[%d]=%q, want=%q", i, targets[i], wantTargets[i])
		}
	}
}
