package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/BurntSushi/toml"
)

const cmd = "bake"

// Task represents a whole task.
type Task struct {
	Command      string
	Args         []string
	Dependencies []string
}

// Command represents a command to run.
type Command struct {
	name string
	args []string
}

var (
	// Command line flags.
	flags       *flag.FlagSet
	showVersion bool
	makeFile    string

	version = "devel"
)

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&makeFile, "f", "bake.toml", "use file as a makefile")
	flags.Usage = usage
}

func main() {
	setFlags()
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", cmd)
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %+v\n", cmd, err)
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) (exitCode int) {
	flags.Parse(args[1:])

	if showVersion {
		fmt.Fprintf(stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	tasks, err := parse(makeFile)
	if err != nil {
		return msg(err)
	}

	target := "default"
	if len(flags.Args()) > 0 {
		target = flags.Args()[0]
	}

	task, found := tasks[target]
	if !found {
		err := fmt.Errorf("'%s' is not defined", target)
		return msg(err)
	}

	commands, err := buildCommands(task, tasks)
	if err != nil {
		return msg(err)
	}

	if err = executeCommands(commands, stdout); err != nil {
		return msg(err)
	}

	return 0
}

func parse(makeFile string) (map[string]Task, error) {
	var p toml.Primitive

	md, err := toml.DecodeFile(makeFile, &p)
	if err != nil {
		return nil, err
	}

	tasks := map[string]Task{}
	if err := md.PrimitiveDecode(p, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func buildCommands(task Task, tasks map[string]Task) ([]Command, error) {
	dependencies := task.Dependencies
	commands := []Command{}

	for len(dependencies) > 0 {
		dependency := dependencies[0]
		dependencies = dependencies[1:]

		t, found := tasks[dependency]
		if !found {
			err := fmt.Errorf("'%s' is not defined", dependency)
			return nil, err
		}

		if len(t.Command) > 0 {
			commands = append([]Command{Command{name: t.Command, args: t.Args}}, commands...)
		}

		dependencies = append(dependencies, t.Dependencies...)
	}

	if len(task.Command) > 0 {
		commands = append(commands, Command{name: task.Command, args: task.Args})
	}

	return commands, nil
}

func executeCommands(commands []Command, stdout io.Writer) error {
	for _, command := range commands {
		out, err := exec.Command(command.name, command.args...).CombinedOutput()
		if err != nil {
			return err
		}

		if len(string(out)) > 0 {
			fmt.Fprintf(stdout, "%s", string(out))
		}
	}

	return nil
}
