package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-xargs"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const (
	flagMaxArgs  = "max-args"
	flagReplace  = "replace"
	flagNull     = "null"
	flagMaxProcs = "max-procs"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `xargs [OPTIONS] [COMMAND [INITIAL-ARGS...]]

Read items from standard input and build command lines to run COMMAND with
those items appended as arguments. With no COMMAND, regroup the input: split
each line into whitespace-separated fields and emit them, at most MAX-ARGS
per output line (default one field per line).`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the xargs CLI against the injected version, I/O, and
// filesystem, returning the process exit code. xargs reads its items from
// stdin; the filesystem is injected for a uniform, testable wiring shape.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, _ afero.Fs) int {
	cmd := newCommand(version, stdin, stdout)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "xargs: %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdin io.Reader, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:            "xargs",
		Version:         version,
		Usage:           "build and execute command lines from standard input",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.IntFlag{Name: flagMaxArgs, Aliases: []string{"n"}, Usage: "use at most MAX-ARGS items per command line"},
			&cli.StringFlag{Name: flagReplace, Aliases: []string{"I"}, Usage: "replace REPLACE-STR in INITIAL-ARGS with each input line; one run per line"},
			&cli.BoolFlag{Name: flagNull, Aliases: []string{"0"}, Usage: "items are NUL-separated, not whitespace-separated"},
			&cli.IntFlag{Name: flagMaxProcs, Aliases: []string{"P"}, Usage: "run up to MAX-PROCS command lines concurrently (output stays in input order)"},
		},
		Action: action(stdin, stdout),
	}
}

func action(stdin io.Reader, stdout io.Writer) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		src := gloo.ByteReaderSource([]io.Reader{stdin})
		_, err := gloo.Run(src, gloo.ByteWriteTo(stdout), command.Xargs(arguments(cmd)...))
		return err
	}
}

// arguments assembles the Xargs constructor arguments: the positional command
// template (program plus its initial arguments) as gloo.File values, followed
// by the option values selected via flags.
func arguments(cmd *cli.Command) []any {
	return append(template(cmd), options(cmd)...)
}

// template renders the positional operands as gloo.File command-template
// arguments. The first is the program to run; the rest are its initial
// arguments. Empty when xargs is invoked with no command (regroup mode).
func template(cmd *cli.Command) []any {
	out := make([]any, cmd.NArg())
	for i := range out {
		out[i] = gloo.File(cmd.Args().Get(i))
	}
	return out
}

func options(cmd *cli.Command) []any {
	var opts []any
	if cmd.IsSet(flagMaxArgs) {
		opts = append(opts, command.XargsMaxArgs(cmd.Int(flagMaxArgs)))
	}
	if cmd.IsSet(flagReplace) {
		opts = append(opts, command.XargsReplace(cmd.String(flagReplace)))
	}
	if cmd.Bool(flagNull) {
		opts = append(opts, command.XargsNull(true))
	}
	if cmd.IsSet(flagMaxProcs) {
		opts = append(opts, command.XargsMaxProcs(cmd.Int(flagMaxProcs)))
	}
	return opts
}
