# bake

The `bake` is a simple task runner. The tasks can be defined in a toml file.

## Example

```toml
# bake.toml
[default]
dependencies = ["build"]

[ls]
command = ["ls"]

[test]
command = ["go", "test", "-v"]

[clean]
command = ["go", "clean"]

[build]
command = ["go", "build"]
dependencies = ["clean"]

[lint]
command = ["golangci-lint", "run", "--disable", "errcheck"]

[all]
dependencies = ["lint", "test", "build"]
```

```bash
$ bake ls
# Run `ls` command.
$ bake all
# Run `golangci-lint`, `go test -v`, `go clean` and `go build`.
```

## Usage

```
$ bake --help
Usage: bake [OPTIONS] [TARGET (default "default")]

OPTIONS:
  -T	print the tasks
  -dry-run
    	print the commands that would be executed
  -f string
    	use file as a configuration file (default "bake.toml")
  -v	print version number
  -verbose
    	use verbose output
```

## Configuration

You can define following values in a configuration file.

* command: A command that execute.
* args: Arguments for a command.
* dependencies: Tasks that before running a command.

### Condition

If you want to switch a command according to OS, you can branch a command inside a makefile.

```
[chrome]
{{if eq .OS "windows"}}
command = ["start", "chrome"]
{{else}}
command = ["google-chrome"]
{{end}}
```

### Variable

You can define variables inside a configuration file.

```
{{$binary:="dummy"}}

[build]
command = ["go", "build", "-o", "{{$binary}}"]
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/bake/releases).
