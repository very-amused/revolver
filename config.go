package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"gopkg.in/yaml.v3"
)

// A config entry specifying a method for opening a filetype and its match conditions
type ConfigItem struct {
	// Optional method label
	Label string `yaml:"label"`

	// File conditions (evaluated as a union - one must match)
	MIME     *regexp.Regexp `yaml:"mime"` // File MIME type
	Ext      *regexp.Regexp `yaml:"ext"`  // Filename extension
	Basename *regexp.Regexp `yaml:"name"` // Filename basename

	// System conditions (evaluated as an intersection - all specified must be true)
	HasProg string `yaml:"has"` // Program is in $PATH

	// Flags
	Fork bool `yaml:"fork"` // Whether to fork when running the command
	Term bool `yaml:"term"` // Whether to run the command in a new terminal

	// Command run in ${SHELL:-/bin/sh} to open the file
	Command string `yaml:"cmd"`
}

// Match - Return whether this entry's conditions match a filepath
func (c *ConfigItem) Match(file string) bool {
	// First check system conditions as an intersection
	if len(c.HasProg) > 0 {
		if _, err := exec.LookPath(c.HasProg); err != nil {
			return false
		}
	}

	// Next check file conditions as a union
	if c.MIME != nil {
		if mime, err := mimetype.DetectFile(file); err == nil && c.MIME.MatchString(mime.String()) {
			return true
		}
	}
	if c.Ext != nil && c.Ext.MatchString(filepath.Ext(file)) {
		return true
	}
	if c.Basename != nil && c.Basename.MatchString(filepath.Base(file)) {
		return true
	}

	return false
}

// buildCommand - Build the final command to be passed to the shell
func (c *ConfigItem) buildCommand(files []string) string {
	var sb strings.Builder
	sb.WriteString("set -- '")
	for _, file := range files {
		// Remove any strings containing NUL
		if strings.ContainsRune(file, '\x00') {
			continue
		}
		// Preserve filenames by shell escaping ' -> '\''
		// Thanks to rifle's codebase for figuring this out, it saved me much head to wall contact
		// https://github.com/ranger/ranger/blob/136416c7e2ecc27315fe2354ecadfe09202df7dd/ranger/ext/rifle.py#L352
		sb.WriteString(strings.ReplaceAll(file, "'", "'\\\\''"))
	}
	sb.WriteString("'; ")
	sb.WriteString(c.Command)

	return sb.String()
}

// Return a string in the form label:flags:cmd
func (c *ConfigItem) String() string {
	var sb strings.Builder
	// Label
	sb.WriteString(c.Label)
	sb.WriteRune(':')

	// Flags
	if c.Fork {
		sb.WriteRune('f')
	}
	if c.Term {
		sb.WriteRune('t')
	}
	sb.WriteRune(':')

	// Command
	sb.WriteString(c.Command)

	return sb.String()
}

type Config []*ConfigItem

// Match - Match a file by extension, returning the nth matching config entry
func (c *Config) Match(file string, n uint) (method *ConfigItem) {
	// Return the first matching command
	var nMatch uint
	for _, method := range *c {
		if method.Match(file) {
			if nMatch == n {
				return method
			}
			nMatch++
		}
	}

	return nil
}

// AllMatches - Get a slice of all matching methods for a file
func (c *Config) AllMatches(file string) (matches []*ConfigItem) {
	for _, method := range *c {
		if method.Match(file) {
			matches = append(matches, method)
		}
	}

	return matches
}

var config Config

func init() {
	// Determine config path
	var configDir = os.Getenv("XDG_CONFIG_HOME")
	if len(configDir) == 0 {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	const configName = "revolver.yaml"
	configPath := filepath.Join(configDir, "revolver", configName)

	// Read and decode config file
	f, err := os.Open(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open %s: %v\n", configName, err)
		os.Exit(1)
	}
	if err = yaml.NewDecoder(f).Decode(&config); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode %s: %v\n", configName, err)
		os.Exit(1)
	}
}
