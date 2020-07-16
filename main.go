package main

import (
	"flag"
	"fmt"
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
	showVersion bool
	makeFile    string

	version = "devel"
)

func init() {
	flag.BoolVar(&showVersion, "v", false, "print version number")
	flag.StringVar(&makeFile, "f", "bake.toml", "use file as a makefile")
	flag.Usage = usage
}

func main() {
	os.Exit(run())
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

func run() int {
	flag.Parse()

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	tasks, err := parse(makeFile)
	if err != nil {
		return msg(err)
	}

	target := "default"
	if len(flag.Args()) > 0 {
		target = flag.Args()[0]
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

	if err = executeCommands(commands); err != nil {
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

func executeCommands(commands []Command) error {
	for _, command := range commands {
		out, err := exec.Command(command.name, command.args...).CombinedOutput()
		if err != nil {
			return err
		}

		if len(string(out)) > 0 {
			fmt.Fprintf(os.Stdout, "%s", string(out))
		}
	}

	return nil
}
