package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/informationsea/shellflow/flowscript"
)

type ShellTaskBuilder struct {
	CurrentID           int
	Tasks               []*ShellTask
	MissingCreatorFiles flowscript.StringSet
	WorkflowContent     string
	workflowLogs        WorkflowLogArray
	config              *Configuration
}

func NewShellTaskBuilder() (*ShellTaskBuilder, error) {
	logs, err := CollectLogs(WorkflowLogDir)
	if err != nil {
		return nil, err
	}

	return &ShellTaskBuilder{
		CurrentID:           0,
		Tasks:               make([]*ShellTask, 0),
		MissingCreatorFiles: flowscript.NewStringSet(),
		workflowLogs:        logs,
	}, nil
}

func (b *ShellTaskBuilder) CreateShellTask(lineNum int, line string) (*ShellTask, error) {
	var formattedLine strings.Builder
	dependentFiles := flowscript.NewStringSet()
	creatingFiles := flowscript.NewStringSet()
	conf, err := LoadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("Cannot load configuration: %s", err.Error())
	}

	// extract dependent and creating files
	for {
		inputStart := strings.Index(line, "((")
		outputStart := strings.Index(line, "[[")

		if inputStart < 0 && outputStart < 0 {
			formattedLine.WriteString(line)
			break
		}

		var endStr string
		var startPos int
		if (outputStart < 0 && inputStart >= 0) || (inputStart >= 0 && inputStart < outputStart) {
			endStr = "))"
			startPos = inputStart
		} else if (inputStart < 0 && outputStart >= 0) || (outputStart >= 0 && outputStart < inputStart) {
			endStr = "]]"
			startPos = outputStart
		}

		//fmt.Printf("startPos: %d / %s / %s\n", startPos, line, endStr)

		formattedLine.WriteString(line[0:startPos])
		line = line[startPos:]
		endPos := strings.Index(line, endStr)
		if endPos < 0 {
			return nil, fmt.Errorf("Closing bracket is not found: %s", endStr)
		}

		targetStr := line[2:endPos]
		formattedLine.WriteString(targetStr)
		line = line[endPos+2:]

		var parsedFiles []string
		if strings.ContainsRune(targetStr, '*') || strings.ContainsRune(targetStr, '?') {
			parsedFiles, err = filepath.Glob(targetStr)
		} else {
			parsedFiles = []string{targetStr}
		}

		switch endStr {
		case "))":
			for _, x := range parsedFiles {
				dependentFiles.Add(x)
			}
		case "]]":
			for _, x := range parsedFiles {
				creatingFiles.Add(x)
			}
		}
	}

	// creating task dependency
	skippable := true
	dependentTasks := make(map[int]struct{})
	missingCreatorFiles := flowscript.NewStringSet()
	for _, v := range dependentFiles.Array() {
		found := false
		for i := len(b.Tasks) - 1; i >= 0; i-- {
			task := b.Tasks[i]
			if task.CreatingFiles.Contains(v) {
				dependentTasks[task.ID] = struct{}{}
				found = true
				break
			}
		}
		if !found {
			missingCreatorFiles.Add(v)
		}
	}
	b.MissingCreatorFiles.AddAll(missingCreatorFiles)
	dependentTaskID := make([]int, 0)
	for k := range dependentTasks {
		if !b.Tasks[k-1].ShouldSkip {
			skippable = false
		}
		dependentTaskID = append(dependentTaskID, k)
	}
	sort.Ints(dependentTaskID)

	//fmt.Printf("skippable: %v : %s\n", skippable, formattedLine.String())
	shellScript := formattedLine.String()

	shouldSkip := false
	var reuseLogPath *JobLog
	if skippable {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		job := b.workflowLogs.SearchReusableJob(shellScript, cwd, dependentFiles, creatingFiles)
		if job != nil { // found
			shouldSkip = true
			reuseLogPath = job
		}
	}

	// check config
	commandConf := CommandConfiguration{
		RegExp:    "",
		SGEOption: []string{},
	}
	//fmt.Printf("config: %d\n", len(conf.Command))
	for _, v := range conf.Command {
		r, err := regexp.Compile(v.RegExp)
		if err != nil {
			return nil, fmt.Errorf("Invalid regular expression in configuration: %s", err.Error())
		}
		//fmt.Printf("checking %s = %s\n", v.RegExp, shellScript)
		if r.MatchString(shellScript) {
			//fmt.Printf("Match\n")
			commandConf = v
			break
		}
	}

	b.CurrentID++
	task := ShellTask{
		LineNum:              lineNum,
		ShellScript:          formattedLine.String(),
		ID:                   b.CurrentID,
		DependentFiles:       dependentFiles,
		CreatingFiles:        creatingFiles,
		DependentTaskID:      dependentTaskID,
		ShouldSkip:           shouldSkip,
		ReuseLog:             reuseLogPath,
		CommandConfiguration: commandConf,
	}

	b.Tasks = append(b.Tasks, &task)
	return &task, nil
}

func (b *ShellTaskBuilder) CreateDag() string {
	var builder strings.Builder
	builder.WriteString("digraph shelltask {\n  node [shape=box];\n")
	for _, v := range b.Tasks {
		builder.WriteString(fmt.Sprintf("  task%d [label=%s];\n", v.ID, strconv.Quote(v.ShellScript)))
	}

	for i, v := range b.MissingCreatorFiles.Array() {
		builder.WriteString(fmt.Sprintf("  input%d [label=%s, color=red];\n", i, strconv.Quote(v)))
		for _, v2 := range b.Tasks {
			if v2.DependentFiles.Contains(v) {
				builder.WriteString(fmt.Sprintf("  input%d -> task%d;\n", i, v2.ID))
			}
		}
	}

	for _, v := range b.Tasks {
		for _, x := range v.DependentTaskID {
			files := v.DependentFiles.Intersect(b.Tasks[x-1].CreatingFiles)
			for _, oneFile := range files.Array() {
				builder.WriteString(fmt.Sprintf("  task%d -> task%d [label=%s];\n", x, v.ID, strconv.Quote(oneFile)))
			}
		}
	}

	allCreatedFiles := make(map[string]int)
	allDependentFiles := make(map[string]int)

	for _, v := range b.Tasks {
		for _, one := range v.DependentFiles.Array() {
			allDependentFiles[one] = v.ID
		}
		for _, one := range v.CreatingFiles.Array() {
			allCreatedFiles[one] = v.ID
		}
	}

	outputID := 0
	for k, v := range allCreatedFiles {
		_, ok := allDependentFiles[k]
		if !ok {
			outputID++
			builder.WriteString(fmt.Sprintf("  output%d [label=%s, color=blue];\n", outputID, strconv.Quote(k)))
			builder.WriteString(fmt.Sprintf("  task%d -> output%d;\n", v, outputID))
		}
	}

	builder.WriteString("}\n")
	return builder.String()
}

type ShellTask struct {
	LineNum              int
	ID                   int
	ShellScript          string
	DependentFiles       flowscript.StringSet
	CreatingFiles        flowscript.StringSet
	DependentTaskID      []int
	ShouldSkip           bool
	ReuseLog             *JobLog
	CommandConfiguration CommandConfiguration
}

func (v *ShellTask) String() string {
	return fmt.Sprintf("SellTask{\n  LineNum: %d, ID: %d,\n  ShellScript: %s,\n  DependentFiles: %s,\n  CreatingFiles: %s,\n  DependentTaskID: %d,\n  ShouldSkip: %v,\n  SGEOption: %s\n}", v.LineNum, v.ID, v.ShellScript, v.DependentFiles.Array(), v.CreatingFiles.Array(), v.DependentTaskID, v.ShouldSkip, v.CommandConfiguration.String())
}
