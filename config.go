package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/gabriel-vasile/mimetype"
	"gopkg.in/yaml.v3"
)

// A config entry specifying a method for opening a filetype and its match conditions
type ConfigItem struct {
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
func (c ConfigItem) Match(file string) bool {
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

type Config []*ConfigItem

// Match - Match a file by extension, returning the nth matching config entry
func (c Config) Match(file string, n uint) (method *ConfigItem) {
	// Return the first matching command
	var nMatch uint
	for _, method := range c {
		if method.Match(file) {
			if nMatch == n {
				return method
			}
			nMatch++
		}
	}

	return nil
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
