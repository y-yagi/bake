package main

import (
	"bytes"
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
