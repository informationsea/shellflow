package flowscript

import (
	"fmt"
	"path"
	"strings"
)

type ScriptFunction struct {
	call      ScriptFunctionCall
	name      string
	maxArgNum int
	minArgNum int
}

func (s ScriptFunction) String() string {
	return s.name
}

var BuiltinFunctions = map[string]*ScriptFunction{
	"basename": &ScriptFunction{ScriptFunctionCallBasename, "basename", 1, 2},
	"dirname":  &ScriptFunction{ScriptFunctionCallDirname, "dirname", 1, 1},
	"prefix":   &ScriptFunction{ScriptFunctionCallPrefix, "prefix", 2, 2},
	"zip":      &ScriptFunction{ScriptFunctionCallZip, "zip", 2, 2},
}

type ScriptFunctionCall func(args []Value) (Value, error)

func ScriptFunctionCallBasename(args []Value) (Value, error) {
	if len(args) >= 3 {
		return nil, fmt.Errorf("Too many arguments for basename")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("one or two arguments are required for basename")
	}

	str, err := args[0].AsString()
	if err != nil {
		return nil, err
	}

	base := path.Base(str)

	if len(args) == 2 {
		str2, err2 := args[1].AsString()
		if err2 != nil {
			return nil, err2
		}
		if strings.HasSuffix(base, str2) {
			return StringValue{base[:len(base)-len(str2)]}, nil
		}
	}
	return StringValue{base}, nil
}

func ScriptFunctionCallDirname(args []Value) (Value, error) {
	if len(args) >= 2 {
		return nil, fmt.Errorf("Too many arguments for dirname")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("one argument are required for dirname")
	}

	str, err := args[0].AsString()
	if err != nil {
		return nil, err
	}

	dir := path.Dir(str)

	return StringValue{dir}, nil
}

func ScriptFunctionCallPrefix(args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("two arguments are required for prefix")
	}

	prefix, err := args[0].AsString()
	if err != nil {
		return nil, err
	}

	array, ok := args[1].(ArrayValue)
	if !ok {
		return nil, fmt.Errorf("%s is not array", args[1])
	}

	values := make([]Value, len(array.value))
	for i, v := range args {
		s, e := v.AsString()
		if e != nil {
			return nil, e
		}
		values[i] = StringValue{prefix + s}
	}

	return ArrayValue{values}, nil
}

func ScriptFunctionCallZip(args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("two arguments are required for zip")
	}

	array1, ok := args[0].(ArrayValue)
	if !ok {
		return nil, fmt.Errorf("%s is not array", args[0])
	}

	array2, ok := args[1].(ArrayValue)
	if !ok {
		return nil, fmt.Errorf("%s is not array", args[1])
	}

	var newLength int
	if len(array1.Value()) > len(array2.Value()) {
		newLength = len(array2.Value())
	} else {
		newLength = len(array1.Value())
	}

	values := make([]Value, newLength)
	for i := 0; i < newLength; i++ {
		pair := [...]Value{array1.Value()[i], array2.Value()[i]}
		values[i] = ArrayValue{pair[:]}
	}

	return ArrayValue{values}, nil
}
