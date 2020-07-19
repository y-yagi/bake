# bake

The `bake` is a simple task runner. The tasks can be defined in a toml file.

## Example

```toml
# bake.toml
[default]
dependencies = ["build"]

[ls]
command = "ls"

[test]
command = "go"
args = ["test", "-v"]

[clean]
command = "go"
args = ["clean"]

[build]
command = "go"
args = ["build"]
dependencies = ["clean"]

[lint]
command = "golangci-lint"
args = ["run", "--disable", "errcheck"]

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
  -dry-run
    	print the commands that would be executed
  -f string
    	use file as a makefile (default "bake.toml")
  -v	print version number
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/bake/releases).
