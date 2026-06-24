# yup-xargs

```
NAME:
   xargs - build and execute command lines from standard input

USAGE:
   xargs [OPTIONS] [COMMAND [INITIAL-ARGS...]]

   Read items from standard input and build command lines to run COMMAND with
   those items appended as arguments. With no COMMAND, regroup the input: split
   each line into whitespace-separated fields and emit them, at most MAX-ARGS
   per output line (default one field per line).

VERSION:
   dev

GLOBAL OPTIONS:
   --max-args int, -n int       use at most MAX-ARGS items per command line (default: 0)
   --replace string, -I string  replace REPLACE-STR in INITIAL-ARGS with each input line; one run per line
   --null, -0                   items are NUL-separated, not whitespace-separated
   --max-procs int, -P int      run up to MAX-PROCS command lines concurrently (output stays in input order) (default: 0)
   --help, -h                   show help
   --version                    print version information and exit
```
