package main

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"

	"github.com/BurntSushi/toml"
)

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
			commands = append(commands, Command{name: t.Command[0], args: t.Command[1:], envs: t.Environments})
		}
	}

	if len(task.Command) > 0 {
		commands = append(commands, Command{name: task.Command[0], args: task.Command[1:], envs: task.Environments})
	}

	if len(task.Commands) > 0 {
		for _, c := range task.Commands {
			commands = append(commands, Command{name: c[0], args: c[1:], envs: task.Environments})
		}
	}

	return commands, nil
}
