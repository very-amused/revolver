package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/gabriel-vasile/mimetype"
	"gopkg.in/yaml.v3"
)

// A config entry specifying how to match and open a filetype
type ConfigItem struct {
	// File conditions (evaluated as a union - one must match)
	MIME     *regexp.Regexp `yaml:"mime"` // File MIME type
	Ext      *regexp.Regexp `yaml:"ext"`  // Filename extension
	Basename *regexp.Regexp `yaml:"name"` // Filename basename

	// System conditions (evaluated as an intersection - all specified must be true)
	HasProg string `yaml:"has"` // Program is in $PATH

	// Flags
	Fork bool `yaml:"fork"` // Whether to fork when running the command

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

// Match - Match a file by extension, returning the command associated with the first match
func (c Config) Match(file string) (command string) {
	// Return the first matching command
	for _, entry := range c {
		if entry.Match(file) {
			return entry.Command
		}
	}

	return ""
}

//go:embed revolver.yaml
var rawConfig []byte

var config Config

func init() {
	if err := yaml.Unmarshal(rawConfig, &config); err != nil {
		panic(fmt.Errorf("Failed to unmarshal revolver.yaml: %v", err))
	}
}
