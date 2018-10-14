package flowscript

import (
	"errors"
	"fmt"
	"strings"
)

var emptyEvaluableArray []Evaluable

// Evaluable is interface of evaluable blocks
type Evaluable interface {
	Evaluate(env Environment) (Value, error)
	String() string
	SubEvaluable() []Evaluable
}

type BadEvaluable struct{}

func (x BadEvaluable) Evaluate(env Environment) (Value, error) {
	return nil, errors.New("bad access")
}

func (x BadEvaluable) String() string {
	return "bad access"
}

func (x BadEvaluable) SubEvaluable() []Evaluable {
	return emptyEvaluableArray
}

type ValueEvaluable struct {
	value Value
}

func (x ValueEvaluable) String() string {
	return x.value.String()
}

func (x ValueEvaluable) Evaluate(env Environment) (Value, error) {
	return x.value, nil
}

func (x ValueEvaluable) SubEvaluable() []Evaluable {
	return emptyEvaluableArray
}

// Variable is one of Evaluable value
type Variable struct {
	Name string
}

// Evaluate variable and get value
func (v *Variable) Evaluate(env Environment) (Value, error) {
	return env.Value(v.Name)
}

func (v *Variable) String() string {
	return v.Name
}

func (x Variable) SubEvaluable() []Evaluable {
	return emptyEvaluableArray
}

type ArrayAccess struct {
	Array      Evaluable
	ArrayIndex Evaluable
}

func (v *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", v.Array, v.ArrayIndex)
}

func (v *ArrayAccess) Evaluate(env Environment) (Value, error) {
	array, err := v.Array.Evaluate(env)
	if err != nil {
		return nil, err
	}
	index, err2 := v.ArrayIndex.Evaluate(env)
	if err2 != nil {
		return nil, err2
	}

	if array, ok := array.(ArrayValue); ok {
		if index, ok := index.(IntValue); ok {
			if int64(len(array.value)) <= index.value || index.value < 0 {
				return nil, fmt.Errorf("out of index error %s %d", v.Array, v.ArrayIndex)
			}
			return array.value[index.value], nil
		}
		return nil, fmt.Errorf("%s is not int", v.ArrayIndex)
	}

	if mapValue, ok := array.(MapValue); ok {
		if mapKey, err := index.AsString(); err == nil {
			value, ok := mapValue.value[mapKey]
			if !ok {
				return nil, fmt.Errorf("%s is not found in %s", mapKey, v.Array)
			}
			return value, nil
		}
		return nil, fmt.Errorf("%s is not string", v.ArrayIndex)
	}
	return nil, fmt.Errorf("%s is not array or map", v.Array)
}

func (x *ArrayAccess) SubEvaluable() []Evaluable {
	evaluables := [...]Evaluable{x.Array, x.ArrayIndex}
	return evaluables[:]
}

type AssignExpression struct {
	variable Evaluable
	exp      Evaluable
}

func (v *AssignExpression) String() string {
	return fmt.Sprintf("%s = %s", v.variable, v.exp)
}

func (v *AssignExpression) Evaluate(env Environment) (Value, error) {
	r, e := v.exp.Evaluate(env)
	if e != nil {
		return nil, e
	}

	variable, ok := v.variable.(*Variable)
	if !ok {
		return nil, fmt.Errorf("%s is not variable", v.variable)
	}

	e = env.Assign(variable.Name, r)
	if e != nil {
		return nil, e
	}

	return r, nil
}

func (x *AssignExpression) SubEvaluable() []Evaluable {
	evaluables := [...]Evaluable{x.variable, x.exp}
	return evaluables[:]
}

type JoinedExpression struct {
	exp1 Evaluable
	exp2 Evaluable
}

func (v *JoinedExpression) String() string {
	return fmt.Sprintf("%s; %s", v.exp1, v.exp2)
}

func (v *JoinedExpression) Evaluate(env Environment) (Value, error) {
	_, e := v.exp1.Evaluate(env)
	if e != nil {
		return nil, e
	}
	return v.exp2.Evaluate(env)
}

func (x *JoinedExpression) SubEvaluable() []Evaluable {
	evaluables := [...]Evaluable{x.exp1, x.exp2}
	return evaluables[:]
}

type PlusExpression struct {
	exp1 Evaluable
	exp2 Evaluable
}

func (v *PlusExpression) String() string {
	return fmt.Sprintf("%s + %s", v.exp1, v.exp2)
}

func (v *PlusExpression) Evaluate(env Environment) (Value, error) {
	r1, e := v.exp1.Evaluate(env)
	if e != nil {
		return nil, e
	}
	r2, e := v.exp2.Evaluate(env)
	if e != nil {
		return nil, e
	}

	if i1, ok := r1.(IntValue); ok {
		if i2, ok := r2.(IntValue); ok {
			return IntValue{i1.value + i2.value}, nil
		}
	}

	if s1, err := r1.AsString(); err == nil {
		if s2, err2 := r2.AsString(); err2 == nil {
			return StringValue{s1 + s2}, nil
		}
	}
	return nil, fmt.Errorf("cannot combine %s and %s", v.exp1, v.exp2)
}

func (x *PlusExpression) SubEvaluable() []Evaluable {
	evaluables := [...]Evaluable{x.exp1, x.exp2}
	return evaluables[:]
}

type NumericOperationExpression struct {
	exp1     Evaluable
	exp2     Evaluable
	operator string
}

func (v *NumericOperationExpression) String() string {
	return fmt.Sprintf("%s %s %s", v.exp1, v.operator, v.exp2)
}

func (v *NumericOperationExpression) Evaluate(env Environment) (Value, error) {
	r1, e := v.exp1.Evaluate(env)
	if e != nil {
		return nil, e
	}
	r2, e := v.exp2.Evaluate(env)
	if e != nil {
		return nil, e
	}

	if i1, err := r1.AsInt(); err == nil {
		if i2, err2 := r2.AsInt(); err2 == nil {
			switch v.operator {
			case "+":
				return IntValue{i1 + i2}, nil
			case "-":
				return IntValue{i1 - i2}, nil
			case "*":
				return IntValue{i1 * i2}, nil
			case "/":
				return IntValue{i1 / i2}, nil
			}

		}
	}

	return nil, fmt.Errorf("cannot calculate %s", v.String())
}

func (x *NumericOperationExpression) SubEvaluable() []Evaluable {
	evaluables := [...]Evaluable{x.exp1, x.exp2}
	return evaluables[:]
}

type FunctionCall struct {
	function Evaluable
	args     []Evaluable
}

func (v *FunctionCall) String() string {
	var b strings.Builder
	b.WriteString(v.function.String())
	b.WriteString("(")
	first := true
	for _, x := range v.args {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(x.String())
	}

	b.WriteString(")")
	return b.String()
}

func (v *FunctionCall) Evaluate(env Environment) (Value, error) {
	functionValue, err := v.function.Evaluate(env)
	if err != nil {
		return nil, err
	}

	if function, ok := functionValue.(FunctionValue); ok {
		values := make([]Value, len(v.args))
		for i, x := range v.args {
			ev, ee := x.Evaluate(env)
			if ee != nil {
				return nil, ee
			}
			values[i] = ev
		}

		return function.value.call(values)
	}
	return nil, fmt.Errorf("%s is not function", functionValue)
}

func (x *FunctionCall) SubEvaluable() []Evaluable {
	evaluables := append(x.args, x.function)
	return evaluables[:]
}

type ArrayExpression struct {
	values []Evaluable
}

func (v *ArrayExpression) String() string {
	var b strings.Builder
	b.WriteString("[")
	first := true
	for _, x := range v.values {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(x.String())
	}

	b.WriteString("]")
	return b.String()
}

func (v *ArrayExpression) Evaluate(env Environment) (Value, error) {
	values := make([]Value, len(v.values))
	for i, x := range v.values {
		ev, ee := x.Evaluate(env)
		if ee != nil {
			return nil, ee
		}
		values[i] = ev
	}

	return ArrayValue{values}, nil
}

func (x *ArrayExpression) SubEvaluable() []Evaluable {
	return x.values
}
