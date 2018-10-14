package main

import (
	"os"
	"path"

	"github.com/informationsea/shellflow/flowscript"
)

const WorkflowLogDir = "shellflow-wf"

type Environment struct {
	flowEnvironment flowscript.Environment
	parameters      map[string]interface{}
	workflowRoot    string
	workDir         string
	skipSha         bool
	dryRun          bool
	scriptsOnly     bool
	rerunAll        bool
}

// NewEnvironment creates new Envrionment value
func NewEnvironment() *Environment {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &Environment{
		flowEnvironment: flowscript.NewGlobalEnvironment(),
		parameters:      make(map[string]interface{}),
		workflowRoot:    path.Join(wd, WorkflowLogDir),
		workDir:         wd,
		skipSha:         false,
		dryRun:          false,
		scriptsOnly:     false,
		rerunAll:        false,
	}
}
