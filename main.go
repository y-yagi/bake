package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/y-yagi/bake/internal/log"
	"github.com/y-yagi/goext/arr"
)

const cmd = "bake"

// Task represents a whole task.
type Task struct {
	Command      []string
	Commands     [][]string
	Dependencies []string
	Environments []string
	Stdout       string
	Stderr       string
}

var (
	// Command line flags.
	flags        *flag.FlagSet
	showVersion  bool
	configFile   string
	dryRun       bool
	verbose      bool
	showTasksFlg bool

	logger  *log.BakeLogger
	version = "devel"
)

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&configFile, "f", "bake.toml", "use file as a configuration file")
	flags.BoolVar(&dryRun, "dry-run", false, "print the commands that would be executed")
	flags.BoolVar(&verbose, "verbose", false, "use verbose output")
	flags.BoolVar(&showTasksFlg, "T", false, "print the tasks")
	flags.Usage = usage
}

func main() {
	setFlags()
	logger = log.NewBakeLogger(os.Stdout)
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
	err := flags.Parse(args[1:])
	if err != nil {
		return msg(err, stderr)
	}

	if showVersion {
		fmt.Fprintf(stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	config := NewConfig(configFile)
	err = config.Parse()
	if err != nil {
		return msg(err, stderr)
	}

	if showTasksFlg {
		showTasks(stdout, config.Tasks)
		return 0
	}

	target := "default"
	if len(flags.Args()) > 0 {
		target = flags.Args()[0]
	}

	commands, err := config.BuildCommands(target)
	if err != nil {
		return msg(err, stderr)
	}

	if err = executeCommands(commands, stdout, stderr); err != nil {
		return msg(err, stderr)
	}

	return 0
}

func executeCommands(commands []Command, stdout, stderr io.Writer) error {
	for _, command := range commands {
		if dryRun {
			fmt.Fprintf(stdout, "%s %s\n", command.name, strings.Join(command.args, " "))
			continue
		}

		if verbose {
			logger.Printf("Run", "%s %s\n", command.name, strings.Join(command.args, " "))
		}
		cmd := exec.Command(command.name, command.args...)
		if command.stdout != io.Discard {
			cmd.Stdout = command.stdout
		} else {
			cmd.Stdout = stdout
		}
		if command.stderr != io.Discard {
			cmd.Stderr = command.stderr
		} else {
			cmd.Stderr = stderr
		}

		if len(command.envs) != 0 {
			cmd.Env = append(os.Environ(), command.envs...)
		}
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func showTasks(stdout io.Writer, tasks map[string]Task) {
	keys := make([]string, 0, len(tasks))
	for k := range tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if len(tasks[k].Commands) > 0 {
			summary := ""
			for _, cmd := range tasks[k].Commands {
				summary += fmt.Sprintf("'%s', ", arr.Join(cmd, " "))
			}
			fmt.Fprintf(stdout, "[%s] %s\n", k, strings.TrimRight(summary, ", "))
		} else {
			cmd := tasks[k].Command
			args := ""
			if len(cmd) == 0 {
				cmd = []string{"*no command*"}
			}
			if len(cmd[1:]) > 0 {
				args = strings.Join(tasks[k].Command[1:], " ")
			}
			fmt.Fprintf(stdout, "[%s] %s %s\n", k, cmd[0], args)
		}
	}
}
