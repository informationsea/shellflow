package flowscript

import (
	"fmt"
)

// Environment contains variable values
type Environment interface {
	// Get value of variable
	Value(key string) (Value, error)
	Assign(key string, value Value) error
	// Get parent environment
	ParentEnvironment() Environment
}

// GlobalEnvironment is Global environment container
type GlobalEnvironment struct {
	env map[string]Value
}

func NewGlobalEnvironment() Environment {
	m := make(map[string]Value)

	for k, v := range BuiltinFunctions {
		m[k] = FunctionValue{v}
	}

	var ge Environment = &GlobalEnvironment{m}
	return ge
}

func CreateSubEnvironment(e Environment) Environment {
	m := make(map[string]Value)
	var se Environment = &SubEnvironment{m, e}
	return se
}

func CreateMixedEnvironment(e Environment, m map[string]Value) Environment {
	var se Environment = &SubEnvironment{m, e}
	return se
}

// Value gets real value from global environment
func (ge *GlobalEnvironment) Value(key string) (Value, error) {
	v, ok := ge.env[key]
	if ok {
		return v, nil
	}
	return nil, fmt.Errorf("Unknown variable %s", key)
}

func (ge *GlobalEnvironment) Assign(key string, value Value) error {
	ge.env[key] = value
	return nil
}

func (ge *GlobalEnvironment) ParentEnvironment() Environment {
	return nil
}

type SubEnvironment struct {
	env    map[string]Value
	parent Environment
}

func (se *SubEnvironment) Value(key string) (Value, error) {
	v, ok := se.env[key]
	if ok {
		return v, nil
	}
	return se.parent.Value(key)
}

func (se *SubEnvironment) Assign(key string, value Value) error {
	se.env[key] = value
	return nil
}

func (se *SubEnvironment) ParentEnvironment() Environment {
	return se.parent
}
