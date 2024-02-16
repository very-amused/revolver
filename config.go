package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"regexp"

	"gopkg.in/yaml.v3"
)

/*
A config entry specifying how to open a file based on the following conditions:

File:
  - MIME type (mime)
  - Extension (ext)
  - Basename (name)

System:
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
	// First check system conditions
	if len(c.HasProg) > 0 {
		if _, err := exec.LookPath(c.HasProg); err != nil {
			return false
		}
	}

	// Next check file conditions
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
