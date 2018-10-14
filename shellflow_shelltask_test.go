package main

import (
	"reflect"
	"testing"

	"github.com/informationsea/shellflow/flowscript"
)

func TestShellTask(t *testing.T) {
	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	{
		shellTask, err := builder.CreateShellTask(1, "java float command")
		if err != nil {
			t.Fatalf("Failed to create shell task %s", err.Error())
		}

		if !reflect.DeepEqual(shellTask, &ShellTask{
			LineNum:              1,
			ID:                   1,
			ShellScript:          "java float command",
			DependentFiles:       flowscript.NewStringSet(),
			CreatingFiles:        flowscript.NewStringSet(),
			DependentTaskID:      []int{},
			CommandConfiguration: CommandConfiguration{RegExp: "java .*", SGEOption: []string{"-l", "s_vmem=40G,mem_req=40G"}},
		}) {
			t.Fatalf("Invalid shell task: %s", shellTask)
		}
	}

	if len(builder.Tasks) != 1 {
		t.Fatalf("Invalid missing task list: %s", builder.Tasks)
	}
	if !reflect.DeepEqual(builder.MissingCreatorFiles.Array(), []string{}) {
		t.Fatalf("Invalid missing creator files: %s", builder.MissingCreatorFiles.Array())
	}

	{
		shellTask, err := builder.CreateShellTask(2, "mid1 ((input1)) [[middle1]]")
		if err != nil {
			t.Fatalf("Failed to create shell task %s", err.Error())
		}

		expectedx := &ShellTask{
			LineNum:              2,
			ID:                   2,
			ShellScript:          "mid1 input1 middle1",
			DependentFiles:       flowscript.NewStringSetWithValues("input1"),
			CreatingFiles:        flowscript.NewStringSetWithValues("middle1"),
			DependentTaskID:      []int{},
			CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
		}

		if !reflect.DeepEqual(shellTask, expectedx) {
			t.Fatalf("Invalid shell task: %s / expected: %s / %t, %t, %t, %t", shellTask, expectedx, shellTask.ShellScript == expectedx.ShellScript, reflect.DeepEqual(shellTask.DependentFiles, expectedx.DependentFiles), reflect.DeepEqual(shellTask.CreatingFiles, expectedx.CreatingFiles), reflect.DeepEqual(shellTask.DependentTaskID, expectedx.DependentTaskID))
		}
	}

	if len(builder.Tasks) != 2 {
		t.Fatalf("Invalid missing task list: %s", builder.Tasks)
	}
	if !reflect.DeepEqual(builder.MissingCreatorFiles.Array(), []string{"input1"}) {
		t.Fatalf("Invalid missing creator files: %s", builder.MissingCreatorFiles.Array())
	}

	{
		shellTask, err := builder.CreateShellTask(5, "mid2 ((input1)) ((middle1)) [[middle2]]")
		if err != nil {
			t.Fatalf("Failed to create shell task %s", err.Error())
		}

		if !reflect.DeepEqual(shellTask, &ShellTask{
			LineNum:              5,
			ID:                   3,
			ShellScript:          "mid2 input1 middle1 middle2",
			DependentFiles:       flowscript.NewStringSetWithValues("input1", "middle1"),
			CreatingFiles:        flowscript.NewStringSetWithValues("middle2"),
			DependentTaskID:      []int{2},
			CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
		}) {
			t.Fatalf("Invalid shell task: %s", shellTask)
		}
	}

	if len(builder.Tasks) != 3 {
		t.Fatalf("Invalid missing task list: %s", builder.Tasks)
	}
	if !reflect.DeepEqual(builder.MissingCreatorFiles.Array(), []string{"input1"}) {
		t.Fatalf("Invalid missing creator files: %s", builder.MissingCreatorFiles.Array())
	}

	{
		shellTask, err := builder.CreateShellTask(10, "mid3 ((input2)) ((middle1)) ((middle2)) [[output]]")
		if err != nil {
			t.Fatalf("Failed to create shell task %s", err.Error())
		}

		if !reflect.DeepEqual(shellTask, &ShellTask{
			LineNum:              10,
			ID:                   4,
			ShellScript:          "mid3 input2 middle1 middle2 output",
			DependentFiles:       flowscript.NewStringSetWithValues("input2", "middle1", "middle2"),
			CreatingFiles:        flowscript.NewStringSetWithValues("output"),
			DependentTaskID:      []int{2, 3},
			CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
		}) {
			t.Fatalf("Invalid shell task: %s", shellTask)
		}

	}

	if s := builder.CreateDag(); s != `digraph shelltask {
  node [shape=box];
  task1 [label="java float command"];
  task2 [label="mid1 input1 middle1"];
  task3 [label="mid2 input1 middle1 middle2"];
  task4 [label="mid3 input2 middle1 middle2 output"];
  input0 [label="input1", color=red];
  input0 -> task2;
  input0 -> task3;
  input1 [label="input2", color=red];
  input1 -> task4;
  task2 -> task3 [label="middle1"];
  task2 -> task4 [label="middle1"];
  task3 -> task4 [label="middle2"];
  output1 [label="output", color=blue];
  task4 -> output1;
}
` {
		t.Fatalf("bad dag: %s", s)
	}

	if _, err := builder.CreateShellTask(1, "echo hello, ((world"); err == nil || err.Error() != "Closing bracket is not found: ))" {
		t.Fatalf("Invalid error: %s", err)
	}

	if _, err := builder.CreateShellTask(1, "echo hello, [[world"); err == nil || err.Error() != "Closing bracket is not found: ]]" {
		t.Fatalf("Invalid error: %s", err)
	}

	if _, err := builder.CreateShellTask(1, "echo ((hello)., ((world"); err == nil || err.Error() != "Closing bracket is not found: ))" {
		t.Fatalf("Invalid error: %s", err)
	}

	if _, err := builder.CreateShellTask(1, "echo [[hello]], [[world"); err == nil || err.Error() != "Closing bracket is not found: ]]" {
		t.Fatalf("Invalid error: %s", err)
	}
}
