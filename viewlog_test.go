package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/informationsea/shellflow/flowscript"
)

func TestCollectLog(t *testing.T) {
	ClearCache()
	tmp, err := NewTempDir("viewlog")
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	shellflow := path.Join(tmp.originalCwd, "shellflow")
	os.Args[0] = shellflow

	defer tmp.Close()

	err = os.Chdir("examples")
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	env := NewEnvironment()
	param := make(map[string]interface{})
	builder, err := parse(env, "build.sf", param)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	gen, err := GenerateTaskScripts("build.sf", "", env, builder)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	// -- before run
	log, err := CollectLogsForOneWork(gen.workflowRoot)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	//fmt.Printf("%s\n", log)

	workflowLogRoot := Abs(path.Join(WorkflowLogDir, path.Base(gen.workflowRoot)))
	expectedLogs := &WorkflowLog{
		WorkflowScript:  Abs("build.sf"),
		WorkflowLogRoot: workflowLogRoot,
		ParameterFile:   "",
		ChangedInput:    []string{},
		StartDate:       log.StartDate,
		JobLogs: []*JobLog{
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job001"),
				IsStarted:          false,
				IsAnyInputChanged:  false,
				IsDone:             false,
				IsAnyOutputChanged: false,
				ExitCode:           -1,
				ScriptExitCode:     -1,
				ShellTask:          builder.Tasks[0],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job002"),
				IsStarted:          false,
				IsAnyInputChanged:  false,
				IsDone:             false,
				IsAnyOutputChanged: false,
				ExitCode:           -1,
				ScriptExitCode:     -1,
				ShellTask:          builder.Tasks[1],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job003"),
				IsStarted:          false,
				IsAnyInputChanged:  false,
				IsDone:             false,
				IsAnyOutputChanged: false,
				ExitCode:           -1,
				ScriptExitCode:     -1,
				ShellTask:          builder.Tasks[2],
			},
		},
	}

	for i, v := range log.JobLogs {
		if !reflect.DeepEqual(v.JobLogRoot, expectedLogs.JobLogs[i].JobLogRoot) {
			t.Fatalf("bad log data[%d] JobLogRoot", i)
		}

		if !reflect.DeepEqual(v.InputFiles, expectedLogs.JobLogs[i].InputFiles) {
			t.Fatalf("bad log data[%d] InputFiles", i)
		}

		if !reflect.DeepEqual(v.OutputFiles, expectedLogs.JobLogs[i].OutputFiles) {
			t.Fatalf("bad log data[%d] output files", i)
		}

		if !reflect.DeepEqual(v.ShellTask, expectedLogs.JobLogs[i].ShellTask) {
			t.Fatalf("bad log data[%d] shell task", i)
		}

		if !reflect.DeepEqual(v, expectedLogs.JobLogs[i]) {

			logsJson, e := json.MarshalIndent(v, "", "  ")
			if e != nil {
				t.Fatalf("error: %s", e.Error())
			}
			expectedLogsJson, e := json.MarshalIndent(expectedLogs.JobLogs[i], "", "  ")
			if e != nil {
				t.Fatalf("error: %s", e.Error())
			}

			t.Fatalf("bad log data[%d]: %s / expected: %s / %v / %v", i, logsJson, expectedLogsJson, reflect.DeepEqual(v.ShellTask, expectedLogs.JobLogs[i].ShellTask), reflect.DeepEqual(logsJson, expectedLogsJson))
		}
	}

	if !reflect.DeepEqual(log, expectedLogs) {
		logsJson, e := json.MarshalIndent(log, "", "  ")
		if e != nil {
			t.Fatalf("error: %s", e.Error())
		}
		expectedLogsJson, e := json.MarshalIndent(expectedLogs, "", "  ")
		if e != nil {
			t.Fatalf("error: %s", e.Error())
		}

		t.Fatalf("Bad log data: %s / expected: %s / %v", logsJson, expectedLogsJson, reflect.DeepEqual(log, &expectedLogs))
	}

	// -------- partial run ----------
	cmd := exec.Command("/bin/bash", "-xe", gen.scripts[1].RunScriptPath)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	err = cmd.Run()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	// -- after one job run
	log, err = CollectLogsForOneWork(gen.workflowRoot)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	//fmt.Printf("After Partial Run: %s\n", logs)

	expectedLogs = &WorkflowLog{
		WorkflowScript:  Abs("build.sf"),
		WorkflowLogRoot: workflowLogRoot,
		ChangedInput:    []string{},
		StartDate:       log.StartDate,
		JobLogs: []*JobLog{
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job001"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[0],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job002"),
				IsStarted:          false,
				IsAnyInputChanged:  false,
				IsDone:             false,
				IsAnyOutputChanged: false,
				ExitCode:           -1,
				ScriptExitCode:     -1,
				ShellTask:          builder.Tasks[1],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job003"),
				IsStarted:          false,
				IsAnyInputChanged:  false,
				IsDone:             false,
				IsAnyOutputChanged: false,
				ExitCode:           -1,
				ScriptExitCode:     -1,
				ShellTask:          builder.Tasks[2],
			},
		},
	}

	if len(log.JobLogs[0].OutputFiles) != 1 {
		t.Fatalf("Invalid number of output files: %s", log)
	}
	expectedLogs.JobLogs[0].OutputFiles = log.JobLogs[0].OutputFiles

	if len(log.JobLogs[0].InputFiles) != 2 {
		t.Fatalf("Invalid number of input files: %s", log)
	}
	expectedLogs.JobLogs[0].InputFiles = log.JobLogs[0].InputFiles

	for j, y := range log.JobLogs {
		if !reflect.DeepEqual(y, expectedLogs.JobLogs[j]) {
			t.Fatalf("Bad log data[%d]: %s / expected: %s", j, y, expectedLogs.JobLogs[j])
		}
	}

	if !reflect.DeepEqual(log, expectedLogs) {
		t.Fatalf("Bad log data: %s / expected: %s", log, expectedLogs)
	}

	// ---------- run all ------------
	time.Sleep(100 * time.Millisecond)
	ClearCache()
	builder, err = parse(env, "build.sf", param)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	gen, err = GenerateTaskScripts("build.sf", "", env, builder)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	//for _, x := range gen.scripts {
	//	fmt.Printf("%s %v\n", x.ScriptPath, x.Skip)
	//}

	workflowLogRoot = path.Join(WorkflowLogDir, path.Base(gen.workflowRoot))

	//for _, x := range builder.Tasks {
	//	fmt.Printf("%s\n", x)
	//}

	err = ExecuteLocalSingle(gen)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	// -- after run
	log, err = CollectLogsForOneWork(workflowLogRoot)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	// fmt.Printf("After Run: %s\n", log)

	expectedLogs = &WorkflowLog{
		WorkflowScript:  Abs("build.sf"),
		WorkflowLogRoot: workflowLogRoot,
		StartDate:       log.StartDate,
		ChangedInput:    []string{},
		JobLogs: []*JobLog{
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job001"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[0],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job002"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[1],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job003"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[2],
			},
		},
	}

	//for _, v := range expectedLog.JobLogs {
	//	fmt.Printf("%s\n", v.ShellTask)
	//}

	for i, v := range log.JobLogs {
		//fmt.Printf("Actual: %s\nExpected: %s\n\n", v.ShellTask, expectedLogs[0].JobLogs[i])

		expectedOutputs := flowscript.NewStringSet()
		for _, u := range v.OutputFiles {
			expectedOutputs.Add(u.Relpath)
		}
		if !reflect.DeepEqual(expectedOutputs.Array(), v.ShellTask.CreatingFiles.Array()) {
			t.Fatalf("bad log: %d : %s", i, v)
		}

		expectedInputs := flowscript.NewStringSet()
		for _, u := range v.InputFiles {
			expectedInputs.Add(u.Relpath)
		}
		if !reflect.DeepEqual(expectedInputs.Array(), v.ShellTask.DependentFiles.Array()) {
			t.Fatalf("bad log: %d : %s", i, v)
		}

		expectedLogs.JobLogs[i].OutputFiles = v.OutputFiles
		expectedLogs.JobLogs[i].InputFiles = v.InputFiles

		if !reflect.DeepEqual(v, expectedLogs.JobLogs[i]) {
			u := expectedLogs.JobLogs[i]

			t.Fatalf(`bad log data: %d: %s / expected: %s`, i, v, u)
		}
	}

	if !reflect.DeepEqual(log, expectedLogs) {
		t.Fatalf("Bad log data: %s / expected: %s", log, expectedLogs)
	}

	// --------- change some file ---------------
	{
		file, err := os.OpenFile("hello.c", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		defer file.Close()
		_, err = file.WriteString("\n")
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
	}

	// clear cache
	ClearCache()

	// -- after changed

	log, err = CollectLogsForOneWork(gen.workflowRoot)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	//fmt.Printf("After Changed: %s\n", logs)

	expectedLogs = &WorkflowLog{
		WorkflowScript:  Abs("build.sf"),
		WorkflowLogRoot: workflowLogRoot,
		ChangedInput:    []string{"hello.c"},
		StartDate:       log.StartDate,
		JobLogs: []*JobLog{
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job001"),
				IsStarted:          true,
				IsAnyInputChanged:  true,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[0],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job002"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[1],
			},
			&JobLog{
				JobLogRoot:         path.Join(workflowLogRoot, "job003"),
				IsStarted:          true,
				IsAnyInputChanged:  false,
				IsDone:             true,
				IsAnyOutputChanged: false,
				ExitCode:           0,
				ScriptExitCode:     0,
				ShellTask:          builder.Tasks[2],
			},
		},
	}

	for i, v := range log.JobLogs {
		expectedOutputs := flowscript.NewStringSet()
		for _, u := range v.OutputFiles {
			expectedOutputs.Add(u.Relpath)
		}
		if !reflect.DeepEqual(expectedOutputs.Array(), v.ShellTask.CreatingFiles.Array()) {
			t.Fatalf("bad log: %d : %s", i, v)
		}

		expectedInputs := flowscript.NewStringSet()
		for _, u := range v.InputFiles {
			expectedInputs.Add(u.Relpath)
		}
		if !reflect.DeepEqual(expectedInputs.Array(), v.ShellTask.DependentFiles.Array()) {
			t.Fatalf("bad log: %d : %s", i, v)
		}

		expectedLogs.JobLogs[i].OutputFiles = v.OutputFiles
		expectedLogs.JobLogs[i].InputFiles = v.InputFiles
	}

	for j, y := range log.JobLogs {
		if !reflect.DeepEqual(y, expectedLogs.JobLogs[j]) {
			t.Fatalf("Bad log data[%d]: %s / expected: %s", j, y, expectedLogs.JobLogs[j])
		}
	}

	if !reflect.DeepEqual(log, expectedLogs) {
		t.Fatalf("Bad log data: %s / expected: %s", log, expectedLogs)
	}

	ViewLog(false, false)
}
