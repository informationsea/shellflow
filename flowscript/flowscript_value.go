package flowscript

import (
	"errors"
	"strconv"
	"strings"
)

// Value representation in flowscript
type Value interface {
	// String representation of value for debug
	String() string
	// Convert to string to embed
	AsString() (string, error)
	// Convert to int
	AsInt() (int64, error)
}

type IntValue struct {
	value int64
}

func NewIntValue(val int64) IntValue {
	return IntValue{val}
}

func (v IntValue) Value() int64 {
	return v.value
}

func (v IntValue) String() string {
	return strconv.FormatInt(v.value, 10)
}

func (v IntValue) AsString() (string, error) {
	return strconv.FormatInt(v.value, 10), nil
}

func (v IntValue) AsInt() (int64, error) {
	return v.value, nil
}

type StringValue struct {
	value string
}

func NewStringValue(str string) StringValue {
	return StringValue{
		value: str,
	}
}

func (v StringValue) Value() string {
	return v.value
}

func (v StringValue) String() string {
	return strconv.Quote(v.value)
}

func (v StringValue) AsString() (string, error) {
	return v.value, nil
}

func (v StringValue) AsInt() (int64, error) {
	return strconv.ParseInt(v.value, 10, 64)
}

type ArrayValue struct {
	value []Value
}

func CreateArrayValue2(v1 Value, v2 Value) ArrayValue {
	array := [...]Value{v1, v2}
	return ArrayValue{array[:]}
}

func (v ArrayValue) Value() []Value {
	return v.value
}

func (v ArrayValue) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, e := range v.value {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(e.String())
	}
	b.WriteString("]")
	return b.String()
}

func (v ArrayValue) AsRawArray() []Value {
	return v.value
}

func (v ArrayValue) AsArray() ([]string, error) {
	array := []string{}
	for _, x := range v.value {
		str, err := x.AsString()
		if err != nil {
			return nil, err
		}
		array = append(array, str)
	}
	return array, nil
}

func (v ArrayValue) AsString() (string, error) {
	var b strings.Builder
	for i, e := range v.value {
		if i != 0 {
			b.WriteString(" ")
		}
		cv, ce := e.AsString()
		if ce != nil {
			return "", ce
		}
		b.WriteString(cv)
	}
	return b.String(), nil
}

func (v ArrayValue) AsInt() (int64, error) {
	return 0, errors.New("Cannot convert array to int")
}

type MapValue struct {
	value map[string]Value
}

func (v MapValue) Value() map[string]Value {
	return v.value
}

func (v MapValue) String() string {
	var b strings.Builder
	first := true
	b.WriteString("{")
	for k, v := range v.value {
		if !first {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v.String())
		first = false
	}
	b.WriteString("}")
	return b.String()
}

func (v MapValue) AsString() (string, error) {
	var b strings.Builder
	first := true
	for k, v := range v.value {
		if !first {
			b.WriteString(" ")
		}
		b.WriteString(k)
		b.WriteString("=")
		cv, ce := v.AsString()
		if ce != nil {
			return "", ce
		}
		b.WriteString(cv)
		first = false
	}
	return b.String(), nil
}

func (v MapValue) AsInt() (int64, error) {
	return 0, errors.New("Cannot convert map to int")
}

type FunctionValue struct {
	value *ScriptFunction
}

func (v FunctionValue) Value() *ScriptFunction {
	return v.value
}

func (v FunctionValue) String() string {
	return v.value.String()
}

func (v FunctionValue) AsString() (string, error) {
	return "", errors.New("Cannot convert function to string")
}

func (v FunctionValue) AsInt() (int64, error) {
	return 0, errors.New("Cannot convert function to int")
}
