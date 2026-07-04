// Command yup-xargs is the CLI wrapper around github.com/gloo-foo/cmd-xargs.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-xargs"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name         = "xargs"
	flagMaxArgs  = "max-args"
	flagReplace  = "replace"
	flagNull     = "null"
	flagMaxProcs = "max-procs"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `xargs [OPTIONS] [COMMAND [INITIAL-ARGS...]]

Read items from standard input and build command lines to run COMMAND with
those items appended as arguments. With no COMMAND, regroup the input: split
each line into whitespace-separated fields and emit them, at most MAX-ARGS
per output line (default one field per line).`

// spec declares the xargs wrapper: a stdin filter whose operands are the
// command template (program plus its initial arguments), configured by flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "build and execute command lines from standard input",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags returns a fresh set of the wrapper's flags. Each call yields new flag
// values, so parsing one invocation never leaks urfave/cli's per-flag "was set"
// state into another (which IsSet reads).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.IntFlag{
			Name:    flagMaxArgs,
			Aliases: []string{"n"},
			Usage:   "use at most MAX-ARGS items per command line",
		},
		&urf.StringFlag{
			Name:    flagReplace,
			Aliases: []string{"I"},
			Usage:   "replace REPLACE-STR in INITIAL-ARGS with each input line; one run per line",
		},
		&urf.BoolFlag{
			Name:    flagNull,
			Aliases: []string{"0"},
			Usage:   "items are NUL-separated, not whitespace-separated",
		},
		&urf.IntFlag{
			Name:    flagMaxProcs,
			Aliases: []string{"P"},
			Usage:   "run up to MAX-PROCS command lines concurrently (output stays in input order)",
		},
	}
}

// build maps the invocation to xargs's pipeline: standard input feeds xargs,
// whose command template (program plus initial arguments) is the operands,
// configured by flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.Stdin(inv.Stdin), command.Xargs(arguments(inv.Args)...), nil
}

// arguments assembles the Xargs constructor arguments: the positional command
// template as file values, followed by the option values selected via flags.
// The first operand is the program to run; the rest are its initial arguments.
// The template is empty when xargs is invoked with no command (regroup mode).
func arguments(c *urf.Command) []any {
	opts := options(c)
	args := make([]any, 0, c.NArg()+len(opts))
	for i := range c.NArg() {
		args = append(args, clix.File(c.Args().Get(i)))
	}
	return append(args, opts...)
}

// options folds the parsed flags into xargs's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagMaxArgs) {
		opts = append(opts, command.XargsMaxArgs(c.Int(flagMaxArgs)))
	}
	if c.IsSet(flagReplace) {
		opts = append(opts, command.XargsReplace(c.String(flagReplace)))
	}
	if c.Bool(flagNull) {
		opts = append(opts, command.XargsNull(true))
	}
	if c.IsSet(flagMaxProcs) {
		opts = append(opts, command.XargsMaxProcs(c.Int(flagMaxProcs)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
