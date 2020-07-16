package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	setFlags()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	run([]string{"bake", "-f", "testdata/sample.toml"}, stdout, stderr)

	got := stdout.String()
	want := "clean\nbuild\n"
	if got != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}
