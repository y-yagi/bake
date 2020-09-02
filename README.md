你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
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

If you want to switch a command according to OS, you can branch a command inside a makefile.

```
[chrome]
{{if eq .OS "windows"}}
command = "start"
args = ["chrome"]
{{else}}
command = "google-chrome"
{{end}}
```

### Variable

You can define variables inside a configuration file.

```
{{$binary:="dummy"}}

[build]
command = "go"
args = ["build", "-o", "{{$binary}}"]
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/bake/releases).
