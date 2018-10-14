package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/informationsea/shellflow/flowscript"
)

func TestSingleFlowScriptTask(t *testing.T) {
	var task FlowTask
	var err error

	{
		ge := flowscript.NewGlobalEnvironment()
		builder, err := NewShellTaskBuilder()
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		task, err = NewSingleFlowScriptTask(10, "#% a = 1; b = 2; c = a")
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}

		if task.Line() != 10 {
			t.Fatalf("Bad line number: %d", task.Line())
		}

		err = task.Subscribe(ge, builder)
		if err != nil {
			t.Fatalf("Cannot submit task: %s", err)
		}

		if v, err := ge.Value("a"); err != nil || v.(flowscript.IntValue).Value() != 1 {
			t.Fatalf("Cannot confirm task result: %s", err)
		}

		if v, err := ge.Value("b"); err != nil || v.(flowscript.IntValue).Value() != 2 {
			t.Fatalf("Cannot confirm task result: %s", err)
		}

		if v, err := ge.Value("c"); err != nil || v.(flowscript.IntValue).Value() != 1 {
			t.Fatalf("Cannot confirm task result: %s", err)
		}

		dependentVariables := task.DependentVariables()
		if !reflect.DeepEqual(dependentVariables.Array(), []string{"a"}) {
			t.Fatalf("Bad dependent variables: %s", dependentVariables.Array())
		}

		createdVariables := task.CreatedVariables()
		if !reflect.DeepEqual(createdVariables.Array(), []string{"a", "b", "c"}) {
			t.Fatalf("Bad created variables: %s", createdVariables.Array())
		}
	}

	{
		task, err = NewSingleFlowScriptTask(10, "#% a = !; b = 2; c = a")
		if err == nil || !strings.HasPrefix(err.Error(), "parse error") {
			t.Fatalf("bad parse result: %s", err)
		}
	}

	{
		task, err = NewSingleFlowScriptTask(10, "# a = !; b = 2; c = a")
		if err == nil || err.Error() != "flowscript line should be started with \"#%\"" {
			t.Fatalf("bad parse result: %s", err)
		}
	}
}

func TestSingleShellTask(t *testing.T) {
	var task FlowTask
	var err error

	{
		ge := flowscript.NewGlobalEnvironment()
		_, err = flowscript.EvaluateScript("b = 100; hoge=\"foo\"", ge)
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}

		task, err = NewSingleShellTask(20, "echo {{a = 1}} {{b}} {{hoge}}")
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}

		if task.Line() != 20 {
			t.Fatalf("Bad line number: %d", task.Line())
		}

		expected := [][]int{{5, 14}, {15, 20}, {21, 29}}
		positions := task.(*SingleShellTask).embeddedPositions
		if !reflect.DeepEqual(expected, positions) {
			t.Fatalf("bad positions: %d (%d)", len(positions), positions)
		}

		runScript, err := task.(*SingleShellTask).EvaluatedShell(ge)
		if err != nil || runScript != "echo 1 100 foo" {
			t.Fatalf("bad result: %s / error: %s", runScript, err)
		}

		dependentVariables := task.DependentVariables()
		if !reflect.DeepEqual(dependentVariables.Array(), []string{"b", "hoge"}) {
			t.Fatalf("Bad dependent variables: %s", dependentVariables.Array())
		}

		createdVariables := task.CreatedVariables()
		if !reflect.DeepEqual(createdVariables.Array(), []string{"a"}) {
			t.Fatalf("Bad created variables: %s", createdVariables.Array())
		}
	}
}

func TestParseShellflow(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
#% x = 1
echo hello, world
#% y = 2

# comment

echo {{x}} {{y}} {{z = 3}}
foo {{z}} {{x}}
`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	param := make(map[string]interface{})
	builder, err := ParseShellflow(reader, env, param)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if l := len(builder.Tasks); l != 3 {
		t.Fatalf("Invalid number of tasks: %d", l)
	}

	if x := builder.Tasks[0]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              3,
		ID:                   1,
		ShellScript:          "echo hello, world",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[1]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              8,
		ID:                   2,
		ShellScript:          "echo 1 2 3",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[2]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              9,
		ID:                   3,
		ShellScript:          "foo 3 1",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}
}

func TestParseShellflowBlock(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
#% x = 1
echo {{x}}

for y in a b c; do
    test {{y}} {{x}}
    hoge {{"foo" + y}}
done

echo {{y}}

`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if content != testScript {
		t.Fatalf("Invalid content: %s", content)
	}

	parsedAssign, err := flowscript.ParseScript(" x = 1")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedEcho, err := NewSingleShellTask(3, "echo {{x}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedTest, err := NewSingleShellTask(6, "test {{y}} {{x}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedHoge, err := NewSingleShellTask(7, "hoge {{\"foo\" + y}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedEcho2, err := NewSingleShellTask(10, "echo {{y}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	expected := &SimpleTaskBlock{
		LineNum: 1,
		SubTask: []FlowTask{
			&SingleScriptFlowTask{
				LineNum:   2,
				Script:    " x = 1",
				evaluable: parsedAssign,
			},
			parsedEcho,
			&ForFlowTask{
				LineNum:      5,
				VariableName: "y",
				Items:        []ForItem{StringForItem("a"), StringForItem("b"), StringForItem("c")},
				SubTask:      []FlowTask{parsedTest, parsedHoge},
			},
			parsedEcho2,
		},
	}

	if !reflect.DeepEqual(block, expected) {
		d, e := json.MarshalIndent(block, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}
		t.Fatalf("Bad block: %s", d)
	}

	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	err = block.Subscribe(env.flowEnvironment, builder)
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	if len(builder.Tasks) != 8 {
		t.Fatalf("Invalid number of tasks: %d", len(builder.Tasks))
	}

	if x := builder.Tasks[0]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              3,
		ID:                   1,
		ShellScript:          "echo 1",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[1]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   2,
		ShellScript:          "test a 1",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[2]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              7,
		ID:                   3,
		ShellScript:          "hoge fooa",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[3]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   4,
		ShellScript:          "test b 1",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[4]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              7,
		ID:                   5,
		ShellScript:          "hoge foob",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[5]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   6,
		ShellScript:          "test c 1",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[6]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              7,
		ID:                   7,
		ShellScript:          "hoge fooc",
		DependentFiles:       flowscript.NewStringSetWithValues(),
		CreatingFiles:        flowscript.NewStringSetWithValues(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[7]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              10,
		ID:                   8,
		ShellScript:          "echo c",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}
}

func TestParseShellflowBlock2(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
for y in examples/*.c; do
    test {{y}}
done
`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if content != testScript {
		t.Fatalf("Invalid content: %s", content)
	}

	parsedTest, err := NewSingleShellTask(3, "test {{y}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	expected := &SimpleTaskBlock{
		LineNum: 1,
		SubTask: []FlowTask{
			&ForFlowTask{
				LineNum:      2,
				VariableName: "y",
				Items:        []ForItem{StringForItem("examples/hello.c"), StringForItem("examples/helloprint.c")},
				SubTask:      []FlowTask{parsedTest},
			},
		},
	}

	if !reflect.DeepEqual(block, expected) {
		d, e := json.MarshalIndent(block, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}
		t.Fatalf("Bad block: %s", d)
	}
}

func TestParseShellflowBlock3(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
#% a = ["examples/hello.c", "examples/helloprint.c", [1, "value", 3]]
for y in {{a}}; do
    test (({{y}}))
done
`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if content != testScript {
		t.Fatalf("Invalid content: %s", content)
	}

	assign, err := NewSingleFlowScriptTask(2, "#% a = [\"examples/hello.c\", \"examples/helloprint.c\", [1, \"value\", 3]]")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedTest, err := NewSingleShellTask(4, "test (({{y}}))")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedVal, err := flowscript.ParseScript("a")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	expected := &SimpleTaskBlock{
		LineNum: 1,
		SubTask: []FlowTask{
			assign,
			&ForFlowTask{
				LineNum:      3,
				VariableName: "y",
				Items:        []ForItem{EvaluableForItem{parsedVal}},
				SubTask:      []FlowTask{parsedTest},
			},
		},
	}

	if !reflect.DeepEqual(block, expected) {
		d, e := json.MarshalIndent(block, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}

		d2, e := json.MarshalIndent(expected, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}
		t.Fatalf("Bad block: %s / %s", d, d2)
	}

	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	err = block.Subscribe(env.flowEnvironment, builder)
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	if len(builder.Tasks) != 3 {
		t.Fatalf("Invalid number of tasks: %d", len(builder.Tasks))
	}

	if x := builder.Tasks[0]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              4,
		ID:                   1,
		ShellScript:          "test examples/hello.c",
		DependentFiles:       flowscript.NewStringSetWithValues("examples/hello.c"),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[1]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              4,
		ID:                   2,
		ShellScript:          "test examples/helloprint.c",
		DependentFiles:       flowscript.NewStringSetWithValues("examples/helloprint.c"),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[2]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              4,
		ID:                   3,
		ShellScript:          "test 1 value 3",
		DependentFiles:       flowscript.NewStringSetWithValues("1 value 3"),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}
}

func TestParseShellflowBlock4(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
#% a = [1, 2, 3]
#% b = [4, 5, 6]
for y in {{zip(a, b)}}; do
    test {{y[0]}} / {{y[1]}}
done
`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if content != testScript {
		t.Fatalf("Invalid content: %s", content)
	}

	assign1, err := NewSingleFlowScriptTask(2, "#% a = [1, 2, 3]")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}
	assign2, err := NewSingleFlowScriptTask(3, "#% b = [4, 5, 6]")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedTest, err := NewSingleShellTask(5, "test {{y[0]}} / {{y[1]}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedVal, err := flowscript.ParseScript("zip(a, b)")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	expected := &SimpleTaskBlock{
		LineNum: 1,
		SubTask: []FlowTask{
			assign1,
			assign2,
			&ForFlowTask{
				LineNum:      4,
				VariableName: "y",
				Items:        []ForItem{EvaluableForItem{parsedVal}},
				SubTask:      []FlowTask{parsedTest},
			},
		},
	}

	if !reflect.DeepEqual(block, expected) {
		d, e := json.MarshalIndent(block, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}

		d2, e := json.MarshalIndent(expected, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}
		t.Fatalf("Bad block: %s / %s", d, d2)
	}

	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	err = block.Subscribe(env.flowEnvironment, builder)
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	if len(builder.Tasks) != 3 {
		t.Fatalf("Invalid number of tasks: %d", len(builder.Tasks))
	}

	if x := builder.Tasks[0]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              5,
		ID:                   1,
		ShellScript:          "test 1 / 4",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[1]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              5,
		ID:                   2,
		ShellScript:          "test 2 / 5",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[2]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              5,
		ID:                   3,
		ShellScript:          "test 3 / 6",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}
}

func TestParseShellflowBlock5(t *testing.T) {
	testScript := `#!/usr/bin/shellflow
#% a = [1, 2]
#% b = [4, 5]
for y in {{zip(a, b)}}; do
    for z in {{y}}; do
        test {{z}}
    done
done
`
	reader := strings.NewReader(testScript)
	env := NewEnvironment()
	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	if content != testScript {
		t.Fatalf("Invalid content: %s", content)
	}

	assign1, err := NewSingleFlowScriptTask(2, "#% a = [1, 2]")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}
	assign2, err := NewSingleFlowScriptTask(3, "#% b = [4, 5]")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedTest, err := NewSingleShellTask(6, "test {{z}}")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedVal1, err := flowscript.ParseScript("zip(a, b)")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	parsedVal2, err := flowscript.ParseScript("y")
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}

	expected := &SimpleTaskBlock{
		LineNum: 1,
		SubTask: []FlowTask{
			assign1,
			assign2,
			&ForFlowTask{
				LineNum:      4,
				VariableName: "y",
				Items:        []ForItem{EvaluableForItem{parsedVal1}},
				SubTask: []FlowTask{
					&ForFlowTask{
						LineNum:      5,
						VariableName: "z",
						Items:        []ForItem{EvaluableForItem{parsedVal2}},
						SubTask:      []FlowTask{parsedTest},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(block, expected) {
		d, e := json.MarshalIndent(block, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}

		d2, e := json.MarshalIndent(expected, "", "  ")
		if e != nil {
			t.Fatalf("bad block and no json: %s", e.Error())
		}
		t.Fatalf("Bad block: %s / %s", d, d2)
	}

	builder, err := NewShellTaskBuilder()
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	err = block.Subscribe(env.flowEnvironment, builder)
	if err != nil {
		t.Fatalf("failed to create shell task builder: %s", err.Error())
	}

	if len(builder.Tasks) != 4 {
		t.Fatalf("Invalid number of tasks: %d", len(builder.Tasks))
	}

	if x := builder.Tasks[0]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   1,
		ShellScript:          "test 1",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[1]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   2,
		ShellScript:          "test 4",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[2]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   3,
		ShellScript:          "test 2",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}

	if x := builder.Tasks[3]; !reflect.DeepEqual(x, &ShellTask{
		LineNum:              6,
		ID:                   4,
		ShellScript:          "test 5",
		DependentFiles:       flowscript.NewStringSet(),
		CreatingFiles:        flowscript.NewStringSet(),
		DependentTaskID:      []int{},
		CommandConfiguration: CommandConfiguration{SGEOption: []string{}},
	}) {
		t.Fatalf("Invalid task:%s", x)
	}
}

func TestForItemSplit(t *testing.T) {
	if x := forItemSplit("a b c"); !reflect.DeepEqual(x, []string{"a", "b", "c"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("  a b c"); !reflect.DeepEqual(x, []string{"a", "b", "c"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("a b c  "); !reflect.DeepEqual(x, []string{"a", "b", "c"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit(" abc efg hij "); !reflect.DeepEqual(x, []string{"abc", "efg", "hij"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("{{[1, 2, 3, 4, 5]}}"); !reflect.DeepEqual(x, []string{"{{[1, 2, 3, 4, 5]}}"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("  {{[1, 2, 3, 4, 5]}}   "); !reflect.DeepEqual(x, []string{"{{[1, 2, 3, 4, 5]}}"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("a  {{[1, 2, 3, 4, 5]}}   c"); !reflect.DeepEqual(x, []string{"a", "{{[1, 2, 3, 4, 5]}}", "c"}) {
		t.Fatalf("bad split result %s", x)
	}

	if x := forItemSplit("a  {{[1, 2, 3, 4, 5] "); !reflect.DeepEqual(x, []string{"a", "{{[1, 2, 3, 4, 5] "}) {
		t.Fatalf("bad split result %s", x)
	}
}
