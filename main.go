package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

const cmd = "bake"

// Task represents a whole task.
type Task struct {
	Command      string
	Args         []string
	Dependencies []string
	Environments []string
}

// Command represents a command to run.
type Command struct {
	name string
	args []string
	envs []string
}

// BakeFileVariable represents a variables of configuration file.
type BakeFileVariable struct {
	OS string
}

var (
	// Command line flags.
	flags       *flag.FlagSet
	showVersion bool
	configFile  string
	dryRun      bool
	verbose     bool

	logger  *BakeLogger
	version = "devel"
)

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&configFile, "f", "bake.toml", "use file as a configuration file")
	flags.BoolVar(&dryRun, "dry-run", false, "print the commands that would be executed")
	flags.BoolVar(&verbose, "verbose", false, "use verbose output")
	flags.Usage = usage
}

func main() {
	setFlags()
	logger = NewLogger(os.Stdout)
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [TARGET (default \"default\")]\n\n", cmd)
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flags.PrintDefaults()
}

func msg(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "%s: %+v\n", cmd, err)
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

	tasks, err := parse(configFile)
	if err != nil {
		return msg(err, stderr)
	}

	target := "default"
	if len(flags.Args()) > 0 {
		target = flags.Args()[0]
	}

	task, found := tasks[target]
	if !found {
		err := fmt.Errorf("'%s' is not defined", target)
		return msg(err, stderr)
	}

	commands, err := buildCommands(task, tasks)
	if err != nil {
		return msg(err, stderr)
	}

	if err = executeCommands(commands, stdout); err != nil {
		return msg(err, stderr)
	}

	return 0
}

func parse(configFile string) (map[string]Task, error) {
	t, err := template.ParseFiles(configFile)
	if err != nil {
		return nil, err
	}

	parsedConfigFile := new(bytes.Buffer)
	tv := BakeFileVariable{OS: runtime.GOOS}
	if err = t.Execute(parsedConfigFile, tv); err != nil {
		return nil, err
	}

	var p toml.Primitive
	md, err := toml.Decode(parsedConfigFile.String(), &p)
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
	definedTasks := map[string]bool{}
	visitedTasks := map[string]bool{}
	commands := []Command{}

	for len(dependencies) > 0 {
		dependency := dependencies[0]

		t, found := tasks[dependency]
		if !found {
			err := fmt.Errorf("'%s' is not defined", dependency)
			return nil, err
		}

		if _, found = definedTasks[dependency]; found {
			err := fmt.Errorf("circular dependency detected, '%s' already added", dependency)
			return nil, err
		}

		if _, found := visitedTasks[dependency]; !found && len(t.Dependencies) > 0 {
			dependencies = append(t.Dependencies, dependencies...)
			visitedTasks[dependency] = true
			continue
		}

		dependencies = dependencies[1:]
		definedTasks[dependency] = true

		if len(t.Command) > 0 {
			commands = append(commands, Command{name: t.Command, args: t.Args, envs: t.Environments})
		}
	}

	if len(task.Command) > 0 {
		commands = append(commands, Command{name: task.Command, args: task.Args, envs: task.Environments})
	}

	return commands, nil
}

func executeCommands(commands []Command, stdout io.Writer) error {
	for _, command := range commands {
		if dryRun {
			fmt.Fprintf(stdout, "%s %s\n", command.name, strings.Join(command.args, " "))
			continue
		}

		if verbose {
			logger.Printf("Run", "%s %s\n", command.name, strings.Join(command.args, " "))
		}
		cmd := exec.Command(command.name, command.args...)
		if len(command.envs) != 0 {
			cmd.Env = append(os.Environ(), command.envs...)
		}
		out, err := cmd.CombinedOutput()
		if len(string(out)) > 0 {
			fmt.Fprintf(stdout, "%s", string(out))
		}
		if err != nil {
			return err
		}
	}

	return nil
}
