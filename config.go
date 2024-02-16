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

/*
A config entry specifying how to open a file based on the following conditions:

File (evaluated as a union - one must be true):
  - MIME type (mime)
  - Extension (ext)
  - Basename (name)

System (evaluated as an intersection - all specified must be true)
  - Has executable in $PATH (has)
*/
type ConfigItem struct {
	MIME     *regexp.Regexp `yaml:"mime"`
	Ext      *regexp.Regexp `yaml:"ext"`
	Basename *regexp.Regexp `yaml:"name"`

	HasProg string `yaml:"has"`

	Command string `yaml:"cmd"`
}

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

type Config []ConfigItem

//go:embed revolver.yaml
var rawConfig []byte

var config Config

func init() {
	if err := yaml.Unmarshal(rawConfig, &config); err != nil {
		panic(fmt.Errorf("Failed to unmarshal revolver.yaml: %v", err))
	}
}
