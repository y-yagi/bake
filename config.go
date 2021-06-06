package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Command represents a command to run.
type Command struct {
	name   string
	args   []string
	envs   []string
	stdout io.Writer
	stderr io.Writer
}

// BakeFileVariable represents a variables of configuration file.
type BakeFileVariable struct {
	OS string
}

// Config represents a configuration file.
type Config struct {
	file  string
	Tasks map[string]Task
}

// NewConfig creates a new Config.
func NewConfig(file string) *Config {
	c := &Config{file: file, Tasks: map[string]Task{}}
	return c
}

// Parse a config file.
func (c *Config) Parse() error {
	t, err := template.ParseFiles(c.file)
	if err != nil {
		return err
	}

	configFile := new(bytes.Buffer)
	tv := BakeFileVariable{OS: runtime.GOOS}
	if err = t.Execute(configFile, tv); err != nil {
		return err
	}

	var p toml.Primitive
	md, err := toml.Decode(configFile.String(), &p)
	if err != nil {
		return err
	}

	if err := md.PrimitiveDecode(p, &c.Tasks); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}

// BuildCommands builds commands.
func (c *Config) BuildCommands(target string) ([]Command, error) {
	task, found := c.Tasks[target]
	if !found {
		err := fmt.Errorf("'%s' is not defined", target)
		return nil, err
	}

	dependencies := task.Dependencies
	definedTasks := map[string]bool{}
	visitedTasks := map[string]bool{}
	commands := []Command{}

	for len(dependencies) > 0 {
		dependency := dependencies[0]

		t, found := c.Tasks[dependency]
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
			stdout, stderr, err := c.openStdoutAndStderr(t.Stdout, t.Stderr)
			if err != nil {
				return nil, err
			}
			commands = append(commands, Command{name: t.Command[0], args: t.Command[1:], envs: t.Environments, stdout: stdout, stderr: stderr})
		}
	}

	if len(task.Command) > 0 {
		stdout, stderr, err := c.openStdoutAndStderr(task.Stdout, task.Stderr)
		if err != nil {
			return nil, err
		}
		commands = append(commands, Command{name: task.Command[0], args: task.Command[1:], envs: task.Environments, stdout: stdout, stderr: stderr})
	}

	if len(task.Commands) > 0 {
		for _, cmd := range task.Commands {
			stdout, stderr, err := c.openStdoutAndStderr(task.Stdout, task.Stderr)
			if err != nil {
				return nil, err
			}
			commands = append(commands, Command{name: cmd[0], args: cmd[1:], envs: task.Environments, stdout: stdout, stderr: stderr})
		}
	}

	return commands, nil
}

func (c *Config) openStdoutAndStderr(stdout string, stderr string) (io.Writer, io.Writer, error) {
	stdoutWriter, stderrWriter := io.Discard, io.Discard
	var err error

	if len(stdout) != 0 {
		stdoutWriter, err = os.OpenFile(stdout, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open 'stdout' path: %v, %v", stdout, err)
		}
	}

	if len(stderr) != 0 {
		stderrWriter, err = os.OpenFile(stderr, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open 'stderr' path: %v, %v", stderr, err)
		}
	}

	return stdoutWriter, stderrWriter, nil
}
