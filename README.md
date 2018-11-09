# goforban
A simple go implementation of forban (https://github.com/adulau/Forban)

## Installation

```
go generate
go install
```

## Usage

```
$ ./goforban                       
goforban commands
=========================

 USAGE: ./goforban <command> [flags]
List of commands:
  help               - print this help
  serve [background] - start forban daemon in foreground
  stop               - stops forban daemon on localhost
  share <file>       - share bundle file
```

`goforban serve` starts forban in the current directory, creating a `var` subdirectory where all runtime files are stored.

Files can be added through `goforban share <filename>` or by copying it to `var/share`.

## Web API

The default web services run on port 12555 using plain HTTP.

Various endpoints exist for interacting with forban:

* `/upload` - POST binary data for sharing, sha checksum will be used as filename
* `/ctrl/stop`- shutdown goforban

Everything under `/ctrl/` is only callable from *127.0.0.1* and *[::1]*.
