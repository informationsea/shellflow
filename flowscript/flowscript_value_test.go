package flowscript

import (
	"testing"
)

type BadType struct{}

func TestValueStringAndType(t *testing.T) {
	{
		var str Value = StringValue{"str"}
		if r, e := str.AsString(); e != nil || r != "str" {
			t.Fatalf("Invalid str: %s / error:%s", r, e)
		}
		if r, e := str.AsInt(); e == nil {
			t.Fatalf("Invalid str: %d / error:%s", r, e)
		}
		if r := str.String(); r != "\"str\"" {
			t.Fatalf("Invalid str: %s", r)
		}
		if v, ok := str.(StringValue); !ok || v.Value() != "str" {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}

	{
		var str Value = StringValue{"456"}
		if r, e := str.AsString(); e != nil || r != "456" {
			t.Fatalf("Invalid str: %s / error:%s", r, e)
		}
		if r, e := str.AsInt(); e != nil || r != 456 {
			t.Fatalf("Invalid str: %d / error:%s", r, e)
		}
		if r := str.String(); r != "\"456\"" {
			t.Fatalf("Invalid str: %s", r)
		}
		if v, ok := str.(StringValue); !ok || v.Value() != "456" {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}

	{
		var num Value = IntValue{123}
		if r, e := num.AsString(); e != nil || r != "123" {
			t.Fatalf("Invalid num: %s / error:%s", r, e)
		}
		if r, e := num.AsInt(); e != nil || r != 123 {
			t.Fatalf("Invalid num: %d / error:%s", r, e)
		}
		if r := num.String(); r != "123" {
			t.Fatalf("Invalid num: %s", r)
		}
		if v, ok := num.(IntValue); !ok || v.Value() != 123 {
			t.Fatalf("Invalid value: %d", v.Value())
		}
	}

	originalArray := [...]Value{IntValue{1}, StringValue{"hoge"}, IntValue{3}}
	{
		var array1 Value = ArrayValue{originalArray[:1]}
		if r, e := array1.AsString(); e != nil || r != "1" {
			t.Fatalf("Invalid array1: %s / error:%s", r, e)
		}
		if r, e := array1.AsInt(); e == nil || e.Error() != "Cannot convert array to int" {
			t.Fatalf("Invalid array1: %d / error:%s", r, e)
		}
		if r := array1.String(); r != "[1]" {
			t.Fatalf("Invalid array1: %s", r)
		}
		if v, ok := array1.(ArrayValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}
	{
		var array2 Value = ArrayValue{originalArray[:2]}
		if r, e := array2.AsString(); e != nil || r != "1 hoge" {
			t.Fatalf("Invalid array2: %s / error:%s", r, e)
		}
		if r, e := array2.AsInt(); e == nil || e.Error() != "Cannot convert array to int" {
			t.Fatalf("Invalid array2: %d / error:%s", r, e)
		}
		if r := array2.String(); r != "[1, \"hoge\"]" {
			t.Fatalf("Invalid array2: %s", r)
		}
		if v, ok := array2.(ArrayValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}
	{
		var array3 Value = ArrayValue{originalArray[:3]}
		if r, e := array3.AsString(); e != nil || r != "1 hoge 3" {
			t.Fatalf("Invalid array3: %s / error:%s", r, e)
		}
		if r, e := array3.AsInt(); e == nil || e.Error() != "Cannot convert array to int" {
			t.Fatalf("Invalid array3: %d / error:%s", r, e)
		}
		if r := array3.String(); r != "[1, \"hoge\", 3]" {
			t.Fatalf("Invalid array3: %s", r)
		}
		if v, ok := array3.(ArrayValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}
	{
		var array4 Value = CreateArrayValue2(IntValue{1}, IntValue{2})
		if r, e := array4.AsString(); e != nil || r != "1 2" {
			t.Fatalf("Invalid array3: %s / error:%s", r, e)
		}
		if r, e := array4.AsInt(); e == nil || e.Error() != "Cannot convert array to int" {
			t.Fatalf("Invalid array3: %d / error:%s", r, e)
		}
		if r := array4.String(); r != "[1, 2]" {
			t.Fatalf("Invalid array3: %s", r)
		}
		if v, ok := array4.(ArrayValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}

	{
		originalMap := make(map[string]Value)
		originalMap["hoge"] = StringValue{"foo"}
		originalMap["foo"] = StringValue{"bar"}
		var mapValue Value = MapValue{originalMap}
		if r, e := mapValue.AsString(); e != nil || (r != "hoge=foo foo=bar" && r != "foo=bar hoge=foo") {
			t.Fatalf("Invalid map: %s / error:%s", r, e)
		}
		if r, e := mapValue.AsInt(); e == nil || e.Error() != "Cannot convert map to int" {
			t.Fatalf("Invalid map: %d / error:%s", r, e)
		}
		if r := mapValue.String(); r != "{hoge=\"foo\", foo=\"bar\"}" && r != "{foo=\"bar\", hoge=\"foo\"}" {
			t.Fatalf("Invalid map: %s", r)
		}
		if v, ok := mapValue.(MapValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}

	{
		basename := BuiltinFunctions["basename"]
		var functionValue Value = FunctionValue{basename}
		if r, e := functionValue.AsString(); e == nil || e.Error() != "Cannot convert function to string" {
			t.Fatalf("Invalid string representation: %s / error:%s", r, e)
		}
		if r, e := functionValue.AsInt(); e == nil || e.Error() != "Cannot convert function to int" {
			t.Fatalf("Invalid functionValue: %d / error:%s", r, e)
		}
		if r := functionValue.String(); r != "basename" {
			t.Fatalf("Invalid functionValue: %s", r)
		}
		if v, ok := functionValue.(FunctionValue); !ok {
			t.Fatalf("Invalid value: %s", v.Value())
		}
	}
}
