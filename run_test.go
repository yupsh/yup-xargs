package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		args       []string
		stdin      string
		wantOut    string
		wantCode   int
		wantErrSub string
	}{
		{
			name:    "default field per line",
			args:    []string{"xargs"},
			stdin:   "one two three\n",
			wantOut: "one\ntwo\nthree\n",
		},
		{
			name:    "max-args groups fields",
			args:    []string{"xargs", "-n", "2"},
			stdin:   "a b c d\n",
			wantOut: "a b\nc d\n",
		},
		{
			name:    "exec runs command once over all fields",
			args:    []string{"xargs", "echo"},
			stdin:   "a b c\n",
			wantOut: "a b c\n",
		},
		{
			name:    "exec groups by max-args",
			args:    []string{"xargs", "-n", "2", "echo"},
			stdin:   "a b c\n",
			wantOut: "a b\nc\n",
		},
		{
			name:    "null-delimited items keep embedded spaces",
			args:    []string{"xargs", "--null", "echo"},
			stdin:   "a b\x00c d\x00",
			wantOut: "a b c d\n",
		},
		{
			name:    "replace token substitutes each line once",
			args:    []string{"xargs", "-I", "{}", "echo", "[{}]"},
			stdin:   "a b\nc\n",
			wantOut: "[a b]\n[c]\n",
		},
		{
			name:    "max-procs preserves input order",
			args:    []string{"xargs", "-P", "4", "-n", "1", "echo"},
			stdin:   "a b c d\n",
			wantOut: "a\nb\nc\nd\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"xargs", "--version"},
			wantOut: "xargs version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"xargs", "--nope"},
			wantCode:   1,
			wantErrSub: "xargs:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, afero.NewMemMapFs())

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
