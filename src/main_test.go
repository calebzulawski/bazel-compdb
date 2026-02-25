package main

import "testing"

func TestReportedVersion(t *testing.T) {
	t.Parallel()

	origModule := moduleVersion
	origBuild := buildVersion
	t.Cleanup(func() {
		moduleVersion = origModule
		buildVersion = origBuild
	})

	tests := []struct {
		name   string
		module string
		build  string
		want   string
	}{
		{
			name:   "module version fallback",
			module: "0.2.0",
			build:  "",
			want:   "0.2.0",
		},
		{
			name:   "build version takes precedence",
			module: "0.0.0",
			build:  "0.2.0",
			want:   "0.2.0",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			moduleVersion = tc.module
			buildVersion = tc.build
			if got := reportedVersion(); got != tc.want {
				t.Fatalf("reportedVersion()=%q, want=%q", got, tc.want)
			}
		})
	}
}
