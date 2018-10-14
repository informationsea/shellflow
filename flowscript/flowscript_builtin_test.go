package flowscript

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type BadValue struct{}

func (v BadValue) Value() interface{} {
	return nil
}

func (v BadValue) String() string {
	return "bad value"
}

func (v BadValue) AsString() (string, error) {
	return "", errors.New("Cannot convert bad value to string")
}

func (v BadValue) AsInt() (int64, error) {
	return 0, errors.New("Cannot convert bad value to int")
}

func TestScriptFunctionCallBasename(t *testing.T) {
	if BuiltinFunctions["basename"].name != "basename" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["basename"].name)
	}

	if BuiltinFunctions["basename"].String() != "basename" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["basename"].name)
	}

	arg1 := [...]Value{StringValue{"a/b/c.c"}, StringValue{".c"}, IntValue{1}}
	if v, e := BuiltinFunctions["basename"].call(arg1[:1]); e != nil || v != (StringValue{"c.c"}) {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["basename"].call(arg1[:2]); e != nil || v != (StringValue{"c"}) {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["basename"].call(arg1[:]); e == nil || e.Error() != "Too many arguments for basename" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["basename"].call(arg1[:0]); e == nil || e.Error() != "one or two arguments are required for basename" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	arg2 := [...]Value{StringValue{"a/b/c.c"}, BadValue{}}
	if v, e := BuiltinFunctions["basename"].call(arg2[:]); e == nil || !strings.HasPrefix(e.Error(), "Cannot convert bad value to string") || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	arg3 := [...]Value{BadValue{}, StringValue{"a/b/c.c"}}
	if v, e := BuiltinFunctions["basename"].call(arg3[:]); e == nil || !strings.HasPrefix(e.Error(), "Cannot convert bad value to string") || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}
}

func TestScriptFunctionCallDirname(t *testing.T) {
	if BuiltinFunctions["dirname"].name != "dirname" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["dirname"].name)
	}

	if BuiltinFunctions["dirname"].String() != "dirname" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["dirname"].name)
	}

	arg1 := [...]Value{StringValue{"a/b/c.c"}, StringValue{".c"}, IntValue{1}}
	if v, e := BuiltinFunctions["dirname"].call(arg1[:1]); e != nil || v != (StringValue{"a/b"}) {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["dirname"].call(arg1[:2]); e == nil || e.Error() != "Too many arguments for dirname" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["dirname"].call(arg1[:0]); e == nil || e.Error() != "one argument are required for dirname" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	arg3 := [...]Value{BadValue{}}
	if v, e := BuiltinFunctions["dirname"].call(arg3[:]); e == nil || !strings.HasPrefix(e.Error(), "Cannot convert bad value to string") || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}
}

func TestScriptFunctionCallPrefix(t *testing.T) {
	if BuiltinFunctions["prefix"].name != "prefix" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["prefix"].name)
	}

	if BuiltinFunctions["prefix"].String() != "prefix" {
		t.Fatalf("Invalid function name: %s", BuiltinFunctions["prefix"].name)
	}

	array := [...]Value{StringValue{"a"}, StringValue{"b"}, StringValue{"c"}}
	arg1 := [...]Value{StringValue{"x"}, ArrayValue{array[:]}, IntValue{1}}
	expected := [...]Value{StringValue{"xa"}, StringValue{"xb"}, StringValue{"xc"}}

	if v, e := BuiltinFunctions["prefix"].call(arg1[:2]); e != nil || reflect.DeepEqual(v, expected) {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["prefix"].call(arg1[:3]); e == nil || e.Error() != "two arguments are required for prefix" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["prefix"].call(arg1[:1]); e == nil || e.Error() != "two arguments are required for prefix" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	if v, e := BuiltinFunctions["prefix"].call(arg1[:0]); e == nil || e.Error() != "two arguments are required for prefix" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	arg3 := [...]Value{IntValue{1}, IntValue{2}}
	if v, e := BuiltinFunctions["prefix"].call(arg3[:]); e == nil || e.Error() != "2 is not array" || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	badArray := [...]Value{BadValue{}, BadValue{}, BadValue{}}
	badArg := [...]Value{StringValue{"hoge"}, ArrayValue{badArray[:]}}
	if v, e := BuiltinFunctions["prefix"].call(badArg[:]); e == nil || !strings.HasPrefix(e.Error(), "Cannot convert bad value to string") || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}

	badArgs2 := [...]Value{BadValue{}, ArrayValue{array[:]}}
	if v, e := BuiltinFunctions["prefix"].call(badArgs2[:]); e == nil || !strings.HasPrefix(e.Error(), "Cannot convert bad value to string") || v != nil {
		t.Fatalf("bad result: %s / error: %s", v, e)
	}
}

func TestScriptFunctionCallZip(t *testing.T) {
	{
		array1 := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}, IntValue{4}}
		array2 := [...]Value{IntValue{5}, IntValue{6}, IntValue{7}}
		args := [...]Value{ArrayValue{array1[:]}, ArrayValue{array2[:]}}
		expectedArray := [...]Value{CreateArrayValue2(IntValue{1}, IntValue{5}), CreateArrayValue2(IntValue{2}, IntValue{6}), CreateArrayValue2(IntValue{3}, IntValue{7})}

		result, err := BuiltinFunctions["zip"].call(args[:])
		if err != nil {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
		resultArray, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
		if !reflect.DeepEqual(resultArray.Value(), expectedArray[:]) {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
	}

	{
		array1 := [...]Value{IntValue{5}, IntValue{6}, IntValue{7}}
		array2 := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}, IntValue{4}}
		args := [...]Value{ArrayValue{array1[:]}, ArrayValue{array2[:]}}
		expectedArray := [...]Value{CreateArrayValue2(IntValue{5}, IntValue{1}), CreateArrayValue2(IntValue{6}, IntValue{2}), CreateArrayValue2(IntValue{7}, IntValue{3})}

		result, err := BuiltinFunctions["zip"].call(args[:])
		if err != nil {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
		resultArray, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
		if !reflect.DeepEqual(resultArray.Value(), expectedArray[:]) {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
	}

	{
		array1 := [...]Value{IntValue{5}, IntValue{6}, IntValue{7}}
		args := [...]Value{ArrayValue{array1[:]}, IntValue{1}}

		result, err := BuiltinFunctions["zip"].call(args[:])
		if err == nil || err.Error() != "1 is not array" {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
	}

	{
		array1 := [...]Value{IntValue{5}, IntValue{6}, IntValue{7}}
		args := [...]Value{IntValue{1}, ArrayValue{array1[:]}}

		result, err := BuiltinFunctions["zip"].call(args[:])
		if err == nil || err.Error() != "1 is not array" {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
	}

	{
		args := [...]Value{IntValue{1}}

		result, err := BuiltinFunctions["zip"].call(args[:])
		if err == nil || err.Error() != "two arguments are required for zip" {
			t.Fatalf("bad result: %s / error: %s", result, err)
		}
	}
}
