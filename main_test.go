package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml"}, stdout, stderr)

	got := stdout.String()
	want := "clean\nbuild\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestCommandFail(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "success"}, stdout, stderr)

	got := stdout.String()
	want := ""
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	got = stderr.String()
	want = "bake: exec: \"zzz\": executable file not found in $PATH\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestDependencyNotDefined(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "not_defined"}, stdout, stderr)

	got := stdout.String()
	want := ""
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	got = stderr.String()
	want = "bake: 'not_defined_dependency' is not defined\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestCircularDependency(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "self"}, stdout, stderr)

	got := stdout.String()
	want := ""
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	got = stderr.String()
	want = "bake: circular dependency detected, 'self' already added\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestSupportTemplate(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "chrome"}, stdout, stderr)

	got := stdout.String()
	want := "google-chrome\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestDryRun(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "-dry-run", "all"}, stdout, stderr)

	got := stdout.String()
	want := "golangci-lint run --disable errcheck\necho test -v\necho clean\necho build\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestEnvironments(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "env"}, stdout, stderr)

	got := stdout.String()
	want := "FOO=BAR"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestTasks(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "-T"}, stdout, stderr)

	got := stdout.String()
	want := "[all] *no command*"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	want = "[echos] 'echo clean', 'echo build'"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestCommands(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "echos"}, stdout, stderr)

	got := stdout.String()
	want := "clean\nbuild\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestStdout(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "stdout"}, stdout, stderr)

	defer os.Remove("testdata/stdout.log")
	got, err := ioutil.ReadFile("testdata/stdout.log")
	if err != nil {
		t.Fatalf("log reading failed %v\n", err)
	}

	want := "clean\n"
	if string(got) != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestStderr(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	run([]string{"bake", "-f", "testdata/sample.toml", "stderr"}, stdout, stderr)

	defer os.Remove("testdata/stderr.log")
	got, err := ioutil.ReadFile("testdata/stderr.log")
	if err != nil {
		t.Fatalf("log reading failed %v\n", err)
	}

	want := "ls: unrecognized option '--invalid-option'\nTry 'ls --help' for more information.\n"
	if string(got) != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}
