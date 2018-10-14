package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/informationsea/shellflow/flowscript"
)

type FlowTask interface {
	DependentVariables() flowscript.StringSet
	CreatedVariables() flowscript.StringSet
	Line() int
	Subscribe(env flowscript.Environment, builder *ShellTaskBuilder) error
}

type FlowTaskBlock interface {
	FlowTask
	AddTask(FlowTask)
}

type SingleScriptFlowTask struct {
	LineNum   int
	Script    string
	evaluable flowscript.Evaluable
}

func NewSingleFlowScriptTask(lineNum int, line string) (*SingleScriptFlowTask, error) {
	if !strings.HasPrefix(line, "#%") {
		return nil, errors.New("flowscript line should be started with \"#%\"")
	}
	parsed, err := flowscript.ParseScript(line[2:])
	if err != nil {
		return nil, err
	}
	return &SingleScriptFlowTask{LineNum: lineNum, Script: line[2:], evaluable: parsed}, nil
}

func (t *SingleScriptFlowTask) Subscribe(env flowscript.Environment, builder *ShellTaskBuilder) error {
	_, e := t.evaluable.Evaluate(env)
	if e != nil {
		return fmt.Errorf("Parse error at line %d: %s", t.LineNum, e.Error())
	}
	return e
}

func (t *SingleScriptFlowTask) Line() int {
	return t.LineNum
}

func (t *SingleScriptFlowTask) DependentVariables() flowscript.StringSet {
	return flowscript.SearchDependentVariables(t.evaluable)
}

func (t *SingleScriptFlowTask) CreatedVariables() flowscript.StringSet {
	return flowscript.SearchCreatedVariables(t.evaluable)
}

type SimpleTaskBlock struct {
	LineNum int
	SubTask []FlowTask
}

func NewSimpleTaskBlock(lineNum int) *SimpleTaskBlock {
	return &SimpleTaskBlock{
		LineNum: lineNum,
		SubTask: make([]FlowTask, 0),
	}
}

func (v *SimpleTaskBlock) AddTask(t FlowTask) {
	v.SubTask = append(v.SubTask, t)
}

func (v *SimpleTaskBlock) DependentVariables() flowscript.StringSet {
	vals := flowscript.NewStringSet()
	for _, x := range v.SubTask {
		vals.AddAll(x.DependentVariables())
	}
	return vals
}

func (v *SimpleTaskBlock) CreatedVariables() flowscript.StringSet {
	vals := flowscript.NewStringSet()
	for _, x := range v.SubTask {
		vals.AddAll(x.CreatedVariables())
	}
	return vals
}

func (v *SimpleTaskBlock) Line() int {
	return v.LineNum
}

func (v *SimpleTaskBlock) Subscribe(env flowscript.Environment, builder *ShellTaskBuilder) error {
	for _, x := range v.SubTask {
		err := x.Subscribe(env, builder)
		if err != nil {
			return err
		}
	}
	return nil
}

type ForItem interface {
	Values(flowscript.Environment) ([]flowscript.Value, error)
}

type StringForItem string

func (s StringForItem) Values(flowscript.Environment) ([]flowscript.Value, error) {
	return []flowscript.Value{flowscript.NewStringValue(string(s))}, nil
}

type EvaluableForItem struct {
	Evaluable flowscript.Evaluable
}

func (s EvaluableForItem) Values(env flowscript.Environment) ([]flowscript.Value, error) {
	value, err := s.Evaluable.Evaluate(env)
	if err != nil {
		return []flowscript.Value{}, err
	}
	if array, ok := value.(flowscript.ArrayValue); ok {
		return array.AsRawArray(), nil
	}
	return []flowscript.Value{value}, nil
}

type ForFlowTask struct {
	VariableName string
	Items        []ForItem
	LineNum      int
	SubTask      []FlowTask
}

var spaceRegexp = regexp.MustCompile(`\s+`)

func forItemSplit(data string) []string {
	result := make([]string, 0)
	pos := 0
	for {
		spacePos := spaceRegexp.FindStringIndex(data[pos:])
		scriptPos := strings.Index(data[pos:], "{{")
		if spacePos == nil && scriptPos < 0 {
			break
		}
		if spacePos != nil && spacePos[0] == 0 {
			//fmt.Printf("skip head space %s\n", strconv.Quote(data[:spacePos[1]]))
			pos += spacePos[1]
		} else if spacePos != nil && (spacePos[0] < scriptPos || scriptPos < 0) {
			//fmt.Printf("add %s  %d %d\n", strconv.Quote(data[pos:pos+spacePos[0]]), pos, spacePos)
			result = append(result, data[pos:pos+spacePos[0]])
			pos += spacePos[1]
		} else if scriptPos >= 0 && (spacePos == nil || scriptPos < spacePos[0]) {
			//fmt.Printf("add %s  %d %d\n", strconv.Quote(data[pos:pos+scriptPos]), pos, scriptPos)
			if scriptPos > 0 {
				result = append(result, data[pos:pos+scriptPos])
				pos += scriptPos
			}

			scriptEndPos := strings.Index(data[pos:], "}}")
			if scriptEndPos < 0 {
				result = append(result, data[pos:])
				pos = len(data)
				break
			} else {
				result = append(result, data[pos:pos+scriptEndPos+2])
				pos += scriptEndPos + 2
			}
		} else {
			fmt.Printf("bad case %d %d\n", scriptPos, spacePos)
			break
		}
	}
	if pos != len(data) {
		result = append(result, data[pos:])
	}
	return result
}

func NewForFlowTask(variableName string, items string, lineNum int) (*ForFlowTask, error) {

	rawItems := forItemSplit(items)
	processedItems := []ForItem{}

	for _, x := range rawItems {
		if strings.HasPrefix(x, "{{") && strings.HasSuffix(x, "}}") {
			parsed, err := flowscript.ParseScript(x[2 : len(x)-2])
			if err != nil {
				return nil, err
			}
			processedItems = append(processedItems, EvaluableForItem{parsed})
		} else if strings.IndexRune(x, '*') >= 0 || strings.IndexRune(x, '?') >= 0 {
			files, err := filepath.Glob(x)
			if err != nil {
				return nil, err
			}
			for _, y := range files {
				processedItems = append(processedItems, StringForItem(y))
			}

		} else {
			processedItems = append(processedItems, StringForItem(x))
		}
	}

	return &ForFlowTask{
		VariableName: variableName,
		Items:        processedItems,
		LineNum:      lineNum,
		SubTask:      make([]FlowTask, 0),
	}, nil
}

func (v *ForFlowTask) AddTask(t FlowTask) {
	v.SubTask = append(v.SubTask, t)
}

func (v *ForFlowTask) DependentVariables() flowscript.StringSet {
	vals := flowscript.NewStringSet()
	for _, x := range v.SubTask {
		vals.AddAll(x.DependentVariables())
	}
	return vals
}

func (v *ForFlowTask) CreatedVariables() flowscript.StringSet {
	vals := flowscript.NewStringSet()
	for _, x := range v.SubTask {
		vals.AddAll(x.CreatedVariables())
	}
	return vals
}

func (v *ForFlowTask) Line() int {
	return v.LineNum
}

func (v *ForFlowTask) Subscribe(env flowscript.Environment, builder *ShellTaskBuilder) error {
	for _, x := range v.Items {
		values, err := x.Values(env)
		if err != nil {
			return err
		}
		for _, z := range values {
			env.Assign(v.VariableName, z)
			for _, y := range v.SubTask {
				err = y.Subscribe(env, builder)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type SingleShellTask struct {
	LineNum           int
	Script            string
	embeddedPositions [][]int
	evaluables        []flowscript.Evaluable
}

var embeddedFlowScriptBrace = regexp.MustCompile("{{[^}]*}}")

func NewSingleShellTask(lineNum int, line string) (*SingleShellTask, error) {
	positions := embeddedFlowScriptBrace.FindAllStringIndex(line, 100)
	var evaluables []flowscript.Evaluable
	for _, v := range positions {
		sub := line[v[0]+2 : v[1]-1]
		//fmt.Printf("sub: %s\n", sub)
		ev, err := flowscript.ParseScript(sub)
		if err != nil {
			return nil, err
		}
		evaluables = append(evaluables, ev)
	}
	return &SingleShellTask{
		LineNum:           lineNum,
		Script:            line,
		embeddedPositions: positions,
		evaluables:        evaluables,
	}, nil
}

func (t *SingleShellTask) DependentVariables() flowscript.StringSet {
	vars := flowscript.NewStringSet()
	for _, v := range t.evaluables {
		vars.AddAll(flowscript.SearchDependentVariables(v))
	}
	return vars
}

func (t *SingleShellTask) CreatedVariables() flowscript.StringSet {
	vars := flowscript.NewStringSet()
	for _, v := range t.evaluables {
		vars.AddAll(flowscript.SearchCreatedVariables(v))
	}
	return vars
}

func (t *SingleShellTask) EvaluatedShell(env flowscript.Environment) (string, error) {
	var results []flowscript.Value = make([]flowscript.Value, len(t.evaluables))
	for i, v := range t.evaluables {
		val, err := v.Evaluate(env)
		if err != nil {
			return "", err
		}
		results[i] = val
	}

	line := t.Script
	for i := len(results) - 1; i >= 0; i-- {
		s, e := results[i].AsString()
		if e != nil {
			return "", e
		}
		line = line[:t.embeddedPositions[i][0]] + (s) + line[t.embeddedPositions[i][1]:]
	}
	return line, nil
}

func (t *SingleShellTask) Subscribe(env flowscript.Environment, builder *ShellTaskBuilder) error {
	line, e := t.EvaluatedShell(env)
	if e != nil {
		return fmt.Errorf("Parse error at line %d: %s", t.LineNum, e.Error())
	}

	_, e = builder.CreateShellTask(t.LineNum, line)
	if e != nil {
		return fmt.Errorf(" error at line %d: %s", t.LineNum, e.Error())
	}

	return nil
}

func (t *SingleShellTask) Line() int {
	return t.LineNum
}

var forBlockRegexp = regexp.MustCompile(`^for\s+(\w+)\s+in\s+(\S.+?)\s*(;?\s*do\s*)?$`)

func ParseShellflowBlock(reader io.Reader, env *Environment) (FlowTaskBlock, string, error) {
	workflowContent := bytes.NewBuffer(nil)
	newReader := io.TeeReader(reader, workflowContent)

	blockStack := []FlowTaskBlock{NewSimpleTaskBlock(1)}

	scanner := bufio.NewScanner(newReader)
	scanner.Split(bufio.ScanLines)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		var task FlowTask
		var err error

		if strings.HasPrefix(line, "for") {
			submatch := forBlockRegexp.FindStringSubmatch(line)
			if submatch == nil {
				return nil, "", fmt.Errorf("Invalid for statement: %s", line)
			}
			//fmt.Printf("for %s in %s\n", submatch[1], submatch[2])

			forTask, err := NewForFlowTask(submatch[1], submatch[2], lineNum)
			if err != nil {
				return nil, "", err
			}
			blockStack[len(blockStack)-1].AddTask(forTask)
			blockStack = append(blockStack, forTask)
			continue
		} else if strings.HasPrefix(line, "done") {
			if line == "done" {
				blockStack = blockStack[0 : len(blockStack)-1]
			} else {
				return nil, "", fmt.Errorf("Invalid done statment: %s", line)
			}
			continue
		} else if strings.HasPrefix(line, "#%") {
			task, err = NewSingleFlowScriptTask(lineNum, line)
		} else if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		} else {
			task, err = NewSingleShellTask(lineNum, line)
		}

		if err != nil {
			return nil, "", err
		}

		blockStack[len(blockStack)-1].AddTask(task)

	}
	if e := scanner.Err(); e != nil {
		return nil, "", e
	}

	return blockStack[0], string(workflowContent.Bytes()), nil
}

func ParseShellflow(reader io.Reader, env *Environment, param map[string]interface{}) (*ShellTaskBuilder, error) {
	env.parameters = param
	for key, value := range param {
		switch value.(type) {
		case string:
			//fmt.Printf("key = %s   string value = %s\n", key, value)
			env.flowEnvironment.Assign(key, flowscript.NewStringValue(value.(string)))
		case float64:
			//fmt.Printf("key = %s   numeric value = %f\n", key, value)
			floatValue := value.(float64)
			env.flowEnvironment.Assign(key, flowscript.NewIntValue(int64(floatValue)))
		default:
			return nil, fmt.Errorf("Unknown parameter type %s = %s", key, value)
		}
	}

	builder, err := NewShellTaskBuilder()
	if err != nil {
		return nil, err
	}

	block, content, err := ParseShellflowBlock(reader, env)
	if err != nil {
		return nil, err
	}

	err = block.Subscribe(env.flowEnvironment, builder)
	if err != nil {
		return nil, err
	}

	builder.WorkflowContent = content

	return builder, nil
}
