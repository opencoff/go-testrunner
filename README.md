[![GoDoc](https://godoc.org/github.com/opencoff/go-testrunner?status.svg)](https://godoc.org/github.com/opencoff/go-testrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/opencoff/go-testrunner)](https://goreportcard.com/report/github.com/opencoff/go-testrunner)

# go-testrunner - Test harness for script driven tests

## What is it?
A library to simplify writing golang tests that act on files and directories.
This library was written to help simplify testing for
[go-fio](https://github.com/opencoff/go-fio)'s `cmp` module. Many of the commands are specifically 
for the `go-fio` module.

## How do I use it?
This library implements common script commands:

* mkfile: Make a file or dir in the test dir; usage:

    mkfile [options] name [name...]

  The optional arguments are:

    -m X | --min-file-size X    Make files at least as big X
    -M Y | --max-file-size Y    Make files no bigger than Y
    -d | --dir                  Make directories instead of files
    -t WhERE | --target WhERE   Make dirs in 'lhs', 'rhs' or 'both

* touch: Sync the timestamps of all dirs and files to be the same.
  The time of start of test is used as the singular reference
  timestamp.

* mutate: modify one or more files

    mutate name [name ...]

* symlink: make one or more symlinks in lhs or rhs or both

    symlink lhs="newname@oldname new2@old2" rhs="newname@oldname"

  Where `lhs=` and `rhs=` denote the target where this is made


## How can I extend it?
Take a look at [go-fio/cmp](https://github.com/opencoff/go-fio/cmp); this module extends
the commands by adding the `expect` command:

* Implements the `testrunnner.Cmd` interface
* Register the command with testrunner by calling `testrunner.RegisterCommand()` from a
  `init()` function.

## Implementation Notes


## License
GPL v2.0
