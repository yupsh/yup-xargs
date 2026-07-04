#!/bin/sh
# Integration checks for yup-xargs, run inside a Debian container with GNU
# findutils (the real `xargs` reference) and coreutils (`echo`, the target).
#
# Every case pipes the same stdin into both yup-xargs and GNU xargs, runs a
# deterministic target command (echo), and compares stdout byte-for-byte.
#
# parity STDIN ARGS...  — yup-xargs must produce byte-identical output to GNU
#                         `xargs` for the same stdin and arguments.
# assert WANT STDIN ARGS... — yup-xargs must produce WANT exactly (used where
#                         yup-xargs diverges from GNU by design; see cmd-xargs
#                         COMPATIBILITY.md).
set -eu

export LC_ALL=C
fails=0

parity() {
	in=$1
	shift
	ours=$(printf '%s' "$in" | yup-xargs "$@" 2>/dev/null || true)
	gnu=$(printf '%s' "$in" | xargs "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  xargs %s\n' "$*"
	else
		printf 'FAIL  parity  xargs %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	in=$2
	shift 2
	got=$(printf '%s' "$in" | yup-xargs "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  xargs %s\n' "$*"
	else
		printf 'FAIL  assert  xargs %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Default exec: items become arguments to one `echo` invocation.
parity 'a b c
' echo
parity 'one
two
three
' echo

# -n N: at most N items per command line, so echo runs once per group.
parity 'a b c d
' -n 2 echo
parity 'a b c d e
' -n 3 echo

# -I {}: substitute the replace-token with each whole input line, one run/line.
parity 'a b
c
' -I '{}' echo '[{}]'

# --null: items are NUL-separated, so an item may contain spaces. (yup-xargs
# accepts the GNU long option `--null`; its short `-0` is not parsed — see the
# divergence note below and COMPATIBILITY.md.)
parity 'a b'"$(printf '\0')"'c d'"$(printf '\0')" --null echo

# --null with -n: NUL items grouped two per echo line.
parity 'a b'"$(printf '\0')"'c d'"$(printf '\0')"'e'"$(printf '\0')" --null -n 2 echo

# Documented divergence: the short `-0` spelling is not recognized by the cli/v3
# parser (it stops at the first non-alphabetic short flag), so `-0` is taken as
# a positional command and the run fails with no stdout. GNU `xargs -0` works;
# yup-xargs requires the long `--null` form (covered above). Assert the empty
# stdout to pin the behavior. See cmd-xargs COMPATIBILITY.md.
assert "" "a$(printf '\0')b$(printf '\0')" -0 echo

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
