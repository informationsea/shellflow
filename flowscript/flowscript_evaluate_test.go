package flowscript

import (
	"reflect"
	"strings"
	"testing"
)

func createTestGlobalEnvironment() Environment {
	ge := NewGlobalEnvironment()

	ge.Assign("hoge", StringValue{"hoge"})
	ge.Assign("foo", StringValue{"foo"})
	ge.Assign("bar", IntValue{1})
	ge.Assign("v1", StringValue{"1"})

	var a1 []Value = make([]Value, 3)
	a1[0] = IntValue{10}
	a1[1] = IntValue{11}
	a1[2] = StringValue{"12"}
	ge.Assign("array", ArrayValue{a1})

	var m1 map[string]Value = make(map[string]Value)
	m1["1"] = IntValue{1}
	m1["2"] = IntValue{2}
	m1["3"] = StringValue{"3!"}
	ge.Assign("map", MapValue{m1})

	return ge
}

func TestConstantValueAccess(t *testing.T) {
	env := createTestGlobalEnvironment()
	var x1 Evaluable = ValueEvaluable{IntValue{1}}
	if v, e := x1.Evaluate(env); e != nil || v != (IntValue{1}) {
		t.Fatalf("Invalid value %s or error %s", v, e)
	}
	if len(x1.SubEvaluable()) != 0 {
		t.Fatalf("too many sub-evaluable: %s", x1.SubEvaluable())
	}

	var x2 Evaluable = ValueEvaluable{StringValue{"hoge"}}
	if v, e := x2.Evaluate(env); e != nil || v != (StringValue{"hoge"}) {
		t.Fatalf("Invalid value %s or error %s", v, e)
	}

	{
		array := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}}
		var x3 Evaluable = ValueEvaluable{ArrayValue{array[:]}}
		v, e := x3.Evaluate(env)
		if e != nil {
			t.Fatalf("Invalid value %s or error %s", v, e)
		}

		if rv, ok := v.(ArrayValue); !ok || !reflect.DeepEqual(array[:], rv.value) {
			t.Fatalf("Invalid value %s or error %s", v, e)
		}
	}

	var badAccess Evaluable = BadEvaluable{}
	if v, e := badAccess.Evaluate(env); e.Error() != "bad access" || v != nil {
		t.Fatalf("Invalid value %s or error %s", v, e)
	}
}

func TestVariableAccess(t *testing.T) {

	var ge Environment = createTestGlobalEnvironment()

	var evaluable Evaluable = &Variable{"hoge"}
	var v Value
	var e error

	v, e = evaluable.Evaluate(ge)
	if e != nil {
		t.Fatal("Cannot access hoge variable")
	}
	if v != (StringValue{"hoge"}) {
		t.Fatalf("Invalid value %s", v)
	}
	if len(evaluable.SubEvaluable()) != 0 {
		t.Fatalf("too many sub-evaluable: %s", evaluable.SubEvaluable())
	}

	evaluable = &Variable{"hoge"}
	v, e = evaluable.Evaluate(ge)
	if e != nil {
		t.Fatal("Cannot access hoge variable")
	}
	if v != (StringValue{"hoge"}) {
		t.Fatalf("Invalid value %s", v)
	}

	evaluable = &Variable{"unknown"}
	v, e = evaluable.Evaluate(ge)
	if e == nil {
		t.Fatal("unknown variable should not exits in global environment #1")
	}
	if v != nil {
		t.Fatal("unknown variable should not exits in global environment #2")
	}

	ge.Assign("unknown", StringValue{"yes"})
	v, e = evaluable.Evaluate(ge)
	if e != nil {
		t.Fatal("Cannot access hoge variable")
	}
	if v != (StringValue{"yes"}) {
		t.Fatalf("Invalid value %s", v)
	}
}

func TestArrayAccess(t *testing.T) {
	ge := createTestGlobalEnvironment()

	originalArray := make([]Value, 3)
	originalArray[0] = IntValue{1}
	originalArray[1] = StringValue{"hoge"}
	originalArray[2] = StringValue{"foo"}

	// check normal access path
	var arrayAccess1 Evaluable = &ArrayAccess{Array: ValueEvaluable{ArrayValue{originalArray}}, ArrayIndex: ValueEvaluable{IntValue{1}}}
	if v, e := arrayAccess1.Evaluate(ge); v != (StringValue{"hoge"}) || e != nil {
		t.Fatalf("Invalid array access result:%s / error:%s", v, e)
	}

	if s := arrayAccess1.String(); s != "[1, \"hoge\", \"foo\"][1]" {
		t.Fatalf("Bad string representation: %s", s)
	}

	subEvaluable := arrayAccess1.SubEvaluable()
	if len(subEvaluable) != 2 || subEvaluable[1] != (ValueEvaluable{IntValue{1}}) {
		t.Fatalf("bad sub-evaluable: %s", subEvaluable)
	}
	subEvaluable1, ok := subEvaluable[0].(ValueEvaluable).value.(ArrayValue)
	if !ok {
		t.Fatalf("bad sub-evaluable: %s", subEvaluable)
	}
	if !reflect.DeepEqual(subEvaluable1.value, originalArray) {
		t.Fatalf("bad sub-evaluable: %s", subEvaluable1.value)
	}

	ge.Assign("foo", ArrayValue{originalArray})
	ge.Assign("bar", IntValue{0})
	var arrayAccess2 Evaluable = &ArrayAccess{Array: &Variable{"foo"}, ArrayIndex: &Variable{"bar"}}
	if v, e := arrayAccess2.Evaluate(ge); v != (IntValue{1}) || e != nil {
		t.Fatalf("Invalid array access result:%s / error:%s", v, e)
	}

	if s := arrayAccess2.String(); s != "foo[bar]" {
		t.Fatalf("Bad string representation: %s", s)
	}

	// check out of index path
	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{ArrayValue{originalArray}}, ArrayIndex: ValueEvaluable{IntValue{3}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) ||
			!strings.HasPrefix(e.Error(), "out of index error") {

			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{ArrayValue{originalArray}}, ArrayIndex: ValueEvaluable{IntValue{-1}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) ||
			!strings.HasPrefix(e.Error(), "out of index error") {

			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	// check invalid type path
	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{StringValue{"hoge"}}, ArrayIndex: ValueEvaluable{IntValue{0}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) || e == nil ||
			e.Error() != "\"hoge\" is not array or map" {
			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{IntValue{123}}, ArrayIndex: ValueEvaluable{IntValue{0}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) || e == nil ||
			e.Error() != "123 is not array or map" {
			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{ArrayValue{originalArray}}, ArrayIndex: ValueEvaluable{StringValue{"hoge"}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) || e == nil ||
			e.Error() != "\"hoge\" is not int" {
			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	// check evaluation fail path
	{
		arrayAccess := ArrayAccess{Array: ValueEvaluable{ArrayValue{originalArray}}, ArrayIndex: BadEvaluable{}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) || e == nil ||
			e.Error() != "bad access" {
			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}

	{
		arrayAccess := ArrayAccess{Array: BadEvaluable{}, ArrayIndex: ValueEvaluable{IntValue{0}}}
		if v, e := arrayAccess.Evaluate(ge); v != (nil) || e == nil ||
			e.Error() != "bad access" {

			t.Fatalf("Invalid array access result:%s / error:%s", v, e)
		}
	}
}

func TestMapAccess(t *testing.T) {
	ge := createTestGlobalEnvironment()
	{
		var mapAccess Evaluable = &ArrayAccess{Array: &Variable{"map"}, ArrayIndex: ValueEvaluable{StringValue{"3"}}}
		if v, e := mapAccess.Evaluate(ge); v != (StringValue{"3!"}) || e != nil {
			t.Fatalf("Invalid map access result:%s / error:%s", v, e)
		}

		subEvaluable := mapAccess.SubEvaluable()
		if len(subEvaluable) != 2 || subEvaluable[1] != (ValueEvaluable{StringValue{"3"}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
		subEvaluable1, ok := subEvaluable[0].(*Variable)
		if !ok || subEvaluable1.Name != "map" {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
	}
	{
		var mapAccess Evaluable = &ArrayAccess{Array: &Variable{"map"}, ArrayIndex: &Variable{"v1"}}
		if v, e := mapAccess.Evaluate(ge); v != (IntValue{1}) || e != nil {
			t.Fatalf("Invalid map access result:%s / error:%s", v, e)
		}
	}

	{
		var mapAccess Evaluable = &ArrayAccess{Array: &Variable{"map"}, ArrayIndex: ValueEvaluable{StringValue{"bad"}}}
		if v, e := mapAccess.Evaluate(ge); v != nil || e.Error() != "bad is not found in map" {
			t.Fatalf("Invalid map access result:%s / error:%s", v, e)
		}
	}

	// check invalid type path
	{
		var mapAccess Evaluable = &ArrayAccess{Array: &Variable{"map"}, ArrayIndex: ValueEvaluable{BadValue{}}}
		if v, e := mapAccess.Evaluate(ge); v != nil || e.Error() != "bad value is not string" {
			t.Fatalf("Invalid map access result:%s / error:%s", v, e)
		}
	}

	// check evaluation fail path
	{
		var mapAccess Evaluable = &ArrayAccess{Array: &Variable{"map"}, ArrayIndex: BadEvaluable{}}
		if v, e := mapAccess.Evaluate(ge); v != nil || e.Error() != "bad access" {
			t.Fatalf("Invalid map access result:%s / error:%s", v, e)
		}
	}
}

func TestPlusExpression(t *testing.T) {
	ge := createTestGlobalEnvironment()
	{
		var plusExpression Evaluable = &PlusExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := plusExpression.Evaluate(ge); e != nil || v != (IntValue{3}) {
			t.Fatalf("Invalid plus operation value: %s / error: %s", v, e)
		}
		if v := plusExpression.String(); v != "1 + 2" {
			t.Fatalf("bad string representation: %s", v)
		}
		subEvaluable := plusExpression.SubEvaluable()
		if len(subEvaluable) != 2 || subEvaluable[0] != (ValueEvaluable{IntValue{1}}) || subEvaluable[1] != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
	}

	{
		var plusExpression Evaluable = &PlusExpression{exp1: ValueEvaluable{StringValue{"hoge"}}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := plusExpression.Evaluate(ge); e != nil || v != (StringValue{"hoge2"}) {
			t.Fatalf("Invalid plus operation value: %s / error: %s", v, e)
		}
		if v := plusExpression.String(); v != "\"hoge\" + 2" {
			t.Fatalf("bad string representation: %s", v)
		}
	}

	{
		var plusExpression Evaluable = &PlusExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{StringValue{"foo"}}}
		if v, e := plusExpression.Evaluate(ge); e != nil || v != (StringValue{"1foo"}) {
			t.Fatalf("Invalid plus operation value: %s / error: %s", v, e)
		}
		if v := plusExpression.String(); v != "1 + \"foo\"" {
			t.Fatalf("bad string representation: %s", v)
		}
	}

	{
		var plusExpression Evaluable = &PlusExpression{exp1: ValueEvaluable{StringValue{"foo"}}, exp2: ValueEvaluable{StringValue{"bar"}}}
		if v, e := plusExpression.Evaluate(ge); e != nil || v != (StringValue{"foobar"}) {
			t.Fatalf("Invalid plus operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation Evaluable = &PlusExpression{exp1: BadEvaluable{}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation Evaluable = &PlusExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: BadEvaluable{}}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation Evaluable = &PlusExpression{exp1: ValueEvaluable{BadValue{}}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || !strings.HasPrefix(e.Error(), "cannot combine") {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation Evaluable = &PlusExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{BadValue{}}}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || !strings.HasPrefix(e.Error(), "cannot combine") {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

}

func TestNumericOperationExpression(t *testing.T) {
	ge := createTestGlobalEnvironment()
	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{IntValue{2}}, operator: "+"}
		if v, e := numericOperation.Evaluate(ge); e != nil || v != (IntValue{3}) {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
		if v := numericOperation.String(); v != "1 + 2" {
			t.Fatalf("bad string representation: %s", v)
		}
		subEvaluable := numericOperation.SubEvaluable()
		if len(subEvaluable) != 2 || subEvaluable[0] != (ValueEvaluable{IntValue{1}}) || subEvaluable[1] != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{IntValue{2}}, operator: "-"}
		if v, e := numericOperation.Evaluate(ge); e != nil || v != (IntValue{-1}) {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
		if v := numericOperation.String(); v != "1 - 2" {
			t.Fatalf("bad string representation: %s", v)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{2}}, exp2: ValueEvaluable{IntValue{2}}, operator: "*"}
		if v, e := numericOperation.Evaluate(ge); e != nil || v != (IntValue{4}) {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
		if v := numericOperation.String(); v != "2 * 2" {
			t.Fatalf("bad string representation: %s", v)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{4}}, exp2: ValueEvaluable{IntValue{2}}, operator: "/"}
		if v, e := numericOperation.Evaluate(ge); e != nil || v != (IntValue{2}) {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
		if v := numericOperation.String(); v != "4 / 2" {
			t.Fatalf("bad string representation: %s", v)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: BadEvaluable{}, exp2: ValueEvaluable{IntValue{2}}, operator: "/"}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != nil || e.Error() != "bad access" {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: BadEvaluable{}, operator: "/"}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation = &NumericOperationExpression{exp1: ValueEvaluable{StringValue{"foo"}}, exp2: ValueEvaluable{IntValue{2}}, operator: "/"}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || !strings.HasPrefix(e.Error(), "cannot calculate") {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}

	{
		var numericOperation Evaluable = &NumericOperationExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{StringValue{"foo"}}, operator: "/"}
		if v, e := numericOperation.Evaluate(ge); e == nil || v != (nil) || !strings.HasPrefix(e.Error(), "cannot calculate") {
			t.Fatalf("Invalid numeric operation value: %s / error: %s", v, e)
		}
	}
}

func TestJoinedExpression(t *testing.T) {
	ge := createTestGlobalEnvironment()

	{
		var joinedExpression Evaluable = &JoinedExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := joinedExpression.Evaluate(ge); e != nil || v != (IntValue{2}) {
			t.Fatalf("Invalid joined expression value: %s / error: %s", v, e)
		}
		if s := joinedExpression.String(); s != "1; 2" {
			t.Fatalf("bad string representation: %s", s)
		}
		subEvaluable := joinedExpression.SubEvaluable()
		if len(subEvaluable) != 2 || subEvaluable[0] != (ValueEvaluable{IntValue{1}}) || subEvaluable[1] != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
	}

	{
		var joinedExpression Evaluable = &JoinedExpression{exp1: BadEvaluable{}, exp2: ValueEvaluable{IntValue{2}}}
		if v, e := joinedExpression.Evaluate(ge); e == nil || v != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid joined expression value: %s / error: %s", v, e)
		}
		if s := joinedExpression.String(); s != "bad access; 2" {
			t.Fatalf("bad string representation: %s", s)
		}
	}

	{
		var joinedExpression Evaluable = &JoinedExpression{exp1: ValueEvaluable{IntValue{1}}, exp2: BadEvaluable{}}
		if v, e := joinedExpression.Evaluate(ge); e == nil || v != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid joined expression value: %s / error: %s", v, e)
		}
	}
}

func TestAssignExpression(t *testing.T) {
	ge := createTestGlobalEnvironment()

	{
		var assignExpression Evaluable = &AssignExpression{variable: &Variable{"x"}, exp: ValueEvaluable{IntValue{2}}}
		v, e := assignExpression.Evaluate(ge)
		av, ae := ge.Value("x")
		if e != nil || ae != nil || v != (IntValue{2}) || av != (IntValue{2}) {
			t.Fatalf("Invalid joined expression value: %s / error: %s / assigned: %s / env error: %s", v, e, av, ae)
		}
		if s := assignExpression.String(); s != "x = 2" {
			t.Fatalf("bad string representation: %s", s)
		}
		subEvaluable := assignExpression.SubEvaluable()
		if len(subEvaluable) != 2 || subEvaluable[0].(*Variable).Name != "x" || subEvaluable[1] != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}
	}

	{
		var assignExpression Evaluable = &AssignExpression{variable: &Variable{"y"}, exp: ValueEvaluable{StringValue{"hoge"}}}
		v, e := assignExpression.Evaluate(ge)
		av, ae := ge.Value("y")
		if e != nil || ae != nil || v != (StringValue{"hoge"}) || av != (StringValue{"hoge"}) {
			t.Fatalf("Invalid joined expression value: %s / error: %s / assigned: %s / env error: %s", v, e, av, ae)
		}
		if s := assignExpression.String(); s != "y = \"hoge\"" {
			t.Fatalf("bad string representation: %s", s)
		}
	}

	{
		var assignExpression Evaluable = &AssignExpression{variable: &Variable{"hoge"}, exp: BadEvaluable{}}
		v, e := assignExpression.Evaluate(ge)
		av, ae := ge.Value("hoge") // check overwriting
		if e == nil || e.Error() != "bad access" || ae != nil || v != (nil) || av != (StringValue{"hoge"}) {
			t.Fatalf("Invalid joined expression value: %s / error: %s / assigned: %s / env error: %s", v, e, av, ae)
		}

	}

	{
		var assignExpression Evaluable = &AssignExpression{variable: ValueEvaluable{IntValue{1}}, exp: ValueEvaluable{IntValue{2}}}
		v, e := assignExpression.Evaluate(ge)
		if e == nil || e.Error() != "1 is not variable" || v != (nil) {
			t.Fatalf("Invalid joined expression value: %s / error: %s", v, e)
		}
	}
}

func TestFunctionCall(t *testing.T) {
	ge := createTestGlobalEnvironment()

	{
		args := [...]Evaluable{ValueEvaluable{StringValue{"hoge/foo"}}}
		dirname := BuiltinFunctions["dirname"]
		var function Evaluable = &FunctionCall{function: ValueEvaluable{FunctionValue{dirname}}, args: args[:]}

		if subEvaluable := function.SubEvaluable(); len(subEvaluable) != 2 || subEvaluable[0].(ValueEvaluable).value != (StringValue{"hoge/foo"}) || subEvaluable[1] != (ValueEvaluable{FunctionValue{dirname}}) {
			t.Fatalf("bad sub-evaluable: %s", subEvaluable)
		}

		if r, e := function.Evaluate(ge); e != nil || r != (StringValue{"hoge"}) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
		if r := function.String(); r != "dirname(\"hoge/foo\")" {
			t.Fatalf("Invalid result: %s", r)
		}
	}

	{
		args := [...]Evaluable{ValueEvaluable{StringValue{"hoge/foo.c"}}, ValueEvaluable{StringValue{".c"}}}
		basename := BuiltinFunctions["basename"]
		var function Evaluable = &FunctionCall{function: ValueEvaluable{FunctionValue{basename}}, args: args[:]}

		if subEvaluable := function.SubEvaluable(); len(subEvaluable) != 3 || subEvaluable[0].(ValueEvaluable).value != (StringValue{"hoge/foo.c"}) || subEvaluable[1].(ValueEvaluable).value != (StringValue{".c"}) || subEvaluable[2] != (ValueEvaluable{FunctionValue{basename}}) {
			t.Fatalf("bad sub-evaluable: %s / condition %t %t %t", subEvaluable, subEvaluable[0].(ValueEvaluable).value != (StringValue{"hoge/foo.c"}), subEvaluable[1].(ValueEvaluable).value != (StringValue{".c"}), subEvaluable[2] != (ValueEvaluable{FunctionValue{basename}}))
		}
		if r, e := function.Evaluate(ge); e != nil || r != (StringValue{"foo"}) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
		if r := function.String(); r != "basename(\"hoge/foo.c\", \".c\")" {
			t.Fatalf("Invalid result: %s", r)
		}
	}

	{
		args := [...]Evaluable{BadEvaluable{}}
		dirname := BuiltinFunctions["dirname"]
		var function Evaluable = &FunctionCall{function: ValueEvaluable{FunctionValue{dirname}}, args: args[:]}
		if r, e := function.Evaluate(ge); e == nil || e.Error() != "bad access" || r != (nil) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
	}

	{
		args := [...]Evaluable{ValueEvaluable{IntValue{1}}}
		var function Evaluable = &FunctionCall{function: BadEvaluable{}, args: args[:]}
		if r, e := function.Evaluate(ge); e == nil || e.Error() != "bad access" || r != (nil) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
	}

	{
		args := [...]Evaluable{ValueEvaluable{IntValue{1}}}
		var function Evaluable = &FunctionCall{function: ValueEvaluable{IntValue{1}}, args: args[:]}
		if r, e := function.Evaluate(ge); e == nil || e.Error() != "1 is not function" || r != (nil) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
	}
}

func TestArrayExpression(t *testing.T) {
	ge := createTestGlobalEnvironment()
	{
		values := [...]Evaluable{ValueEvaluable{IntValue{1}}, ValueEvaluable{IntValue{2}}, ValueEvaluable{IntValue{3}}}
		expected := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}}
		var arrayExpression Evaluable = &ArrayExpression{values[:]}
		r, e := arrayExpression.Evaluate(ge)
		if e != nil {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
		array, ok := r.(ArrayValue)
		if !ok {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}

		if !reflect.DeepEqual(array.value, expected[:]) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}

		if s := arrayExpression.String(); s != "[1, 2, 3]" {
			t.Fatalf("Invalid string result: %s", s)
		}

		if !reflect.DeepEqual(arrayExpression.SubEvaluable(), values[:]) {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
	}

	{
		values := [...]Evaluable{ValueEvaluable{IntValue{1}}, BadEvaluable{}, ValueEvaluable{IntValue{3}}}
		var arrayExpression Evaluable = &ArrayExpression{values[:]}
		if r, e := arrayExpression.Evaluate(ge); e == nil || r != (nil) || e.Error() != "bad access" {
			t.Fatalf("Invalid result: %s / error: %s", r, e)
		}
	}
}

func TestSearchDependentVariables(t *testing.T) {
	tokenizer := createInitializedTokenizer("a = b + c + d; k = x / y + basename(\"foo/bar\", z) + b; l = m")
	evaluable, err := ParseAsExp(tokenizer)
	if err != nil {
		t.Fatalf("Failed to parse: %s", err)
	}
	dependentVariables := SearchDependentVariables(evaluable).Array()

	expected := []string{"b", "basename", "c", "d", "m", "x", "y", "z"}
	if !reflect.DeepEqual(expected, dependentVariables) {
		t.Fatalf("bad dependent variables: %s", dependentVariables)
	}
}

func TestSearchCreatedVariables(t *testing.T) {
	tokenizer := createInitializedTokenizer("a = b + c + d; k = x / y + basename(\"foo/bar\", z) + b; l = m")
	evaluable, err := ParseAsExp(tokenizer)
	if err != nil {
		t.Fatalf("Failed to parse: %s", err)
	}
	dependentVariables := SearchCreatedVariables(evaluable).Array()

	expected := []string{"a", "k", "l"}
	if !reflect.DeepEqual(expected, dependentVariables) {
		t.Fatalf("bad created variables: %s", dependentVariables)
	}
}
