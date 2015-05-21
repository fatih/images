# Flags [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fatih/flags) [![Build Status](http://img.shields.io/travis/fatih/flags.svg?style=flat-square)](https://travis-ci.org/fatih/flags)


Flags is a low level package for parsing or managing single flag arguments and
their associated values from a list of arguments. It's useful for CLI
applications or creating logic for parsing arguments(custom or os.Args)
manually. Checkout the
usage below for examples:

## Install

```bash
go get github.com/fatih/flags
```

## Usage and examples

Let us define three flags. Flags needs to be compatible with the
[flag](https://golang.org/pkg/flag/) package.

```go
args := []string{"--key", "123", "--name=example", "--debug"}
```

Check if a flag exists in the argument list

```go
flags.Has("key", args)    // true
flags.Has("--key", args)  // true
flags.Has("secret", args) // false
```

Get the value for from a flag name

```go
val, _ := flags.Value("--key", args) // val -> "123"
val, _ := flags.Value("name", args)  // val -> "example"
val, _ := flags.Value("debug", args) // val -> "" (means true boolean)
```

Exclude a flag and it's value from the argument list

```go
rArgs := flags.Exclude("key", args)  // rArgs -> ["--name=example", "--debug"]
rArgs := flags.Exclude("name", args) // rArgs -> ["--key", "123", "--debug"]
rArgs := flags.Exclude("foo", args)  // rArgs -> ["--key", "123", "--name=example "--debug"]
```

Is a flag in its valid representation (compatible with the flag package)?

```go
flags.Valid("foo")      // false
flags.Valid("--foo")    // true
flags.Valid("-key=val") // true
flags.Valid("-name=")   // true
```

Parse a flag and return the name

```go
name, _ := flags.Parse("foo")        // returns error, because foo is invalid
name, _ := flags.Parse("--foo")      // name -> "foo
name, _ := flags.Parse("-foo")       // name -> "foo
name, _ := flags.Parse("-foo=value") // name -> "foo
name, _ := flags.Parse("-foo=")      // name -> "foo
```

