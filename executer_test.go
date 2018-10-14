package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/informationsea/shellflow/flowscript"
)

func TestGenerateTaskScripts(t *testing.T) {
	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	env := NewEnvironment()
	tempdir, err := ioutil.TempDir("", "generate_task_scripts")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %s", tempdir)
	}
	fmt.Printf("generate task scripts: %s\n", tempdir)

	env.workDir = path.Join(tempdir, "workdir")
	env.workflowRoot = path.Join(tempdir, "root")

	err = os.MkdirAll(env.workDir, 0755)
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	err = os.MkdirAll(env.workflowRoot, 0755)
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	err = ioutil.WriteFile(path.Join(env.workDir, "hoge"), []byte("foo"), 0640)
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	builder.MissingCreatorFiles.Add("hoge")
	builder.Tasks = append(builder.Tasks, &ShellTask{
		LineNum:         1,
		ID:              1,
		ShellScript:     "cat hoge > foo",
		DependentFiles:  flowscript.NewStringSetWithValues("hoge"),
		CreatingFiles:   flowscript.NewStringSetWithValues("foo"),
		DependentTaskID: []int{},
	})
	builder.Tasks = append(builder.Tasks, &ShellTask{
		LineNum:         2,
		ID:              2,
		ShellScript:     "cat foo > bar",
		DependentFiles:  flowscript.NewStringSetWithValues("foo"),
		CreatingFiles:   flowscript.NewStringSetWithValues("bar"),
		DependentTaskID: []int{1},
	})
	builder.Tasks = append(builder.Tasks, &ShellTask{
		LineNum:         3,
		ID:              3,
		ShellScript:     "cat foo hoge > bar2",
		DependentFiles:  flowscript.NewStringSetWithValues("hoge", "foo"),
		CreatingFiles:   flowscript.NewStringSetWithValues("bar2"),
		DependentTaskID: []int{1},
	})

	scripts, err := GenerateTaskScripts("testscript.sf", "", env, builder)
	if err != nil {
		t.Fatalf("cannot create task scripts: %s", err.Error())
	}

	for _, oneTask := range builder.Tasks {
		jobPath := path.Join(scripts.workflowRoot, fmt.Sprintf("job%03d", oneTask.ID))
		dirStat, err := os.Stat(jobPath)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		if !dirStat.IsDir() {
			t.Fatal("JOB directory should be directory")
		}

		scriptPath := path.Join(jobPath, "run.sh")
		_, err = os.Stat(scriptPath)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		scriptPath = path.Join(jobPath, "script.sh")
		_, err = os.Stat(scriptPath)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		scriptFile, err := os.Open(scriptPath)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		data, err := ioutil.ReadAll(scriptFile)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}
		if string(data) != oneTask.ShellScript {
			t.Fatalf("bad script: %s", data)
		}
	}

	fmt.Printf("dir: %s\n", scripts.workflowRoot)

	os.RemoveAll(tempdir)
}
