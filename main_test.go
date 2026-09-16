package main

import (
	"context"
	"strings"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  flags(),
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name}, 0},
		{"max-args", []string{name, "-n", "2"}, 1},
		{"replace", []string{name, "-I", "{}"}, 1},
		{"null", []string{name, "--null"}, 1},
		{"max-procs", []string{name, "-P", "4"}, 1},
		{"every", []string{name, "-n", "2", "-I", "{}", "--null", "-P", "4"}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(options(parse(t, tc.args...))); got != tc.want {
				t.Fatalf("options len=%d, want %d", got, tc.want)
			}
		})
	}
}

func TestArguments_TemplateThenOptions(t *testing.T) {
	got := arguments(parse(t, name, "echo", "hi", "-n", "2"))
	if len(got) != 3 { // two template operands + one option
		t.Fatalf("arguments len=%d, want 3", len(got))
	}
	if _, ok := got[0].(clix.File); !ok {
		t.Fatalf("arguments[0] type=%T, want clix.File", got[0])
	}
}

func TestArguments_RegroupMode(t *testing.T) {
	if got := arguments(parse(t, name)); len(got) != 0 {
		t.Fatalf("arguments len=%d, want 0 for no command", len(got))
	}
}

func TestBuild(t *testing.T) {
	inv := clix.Invocation{Args: parse(t, name), Stdin: strings.NewReader(""), Fs: afero.NewMemMapFs()}
	src, filter, err := build(inv)
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}
