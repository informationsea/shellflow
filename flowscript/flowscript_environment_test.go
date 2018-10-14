package flowscript

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	ge := NewGlobalEnvironment()
	if p := ge.ParentEnvironment(); p != nil {
		t.Fatalf("Global Environment should not have parent environment: %s", p)
	}

	{
		v, e := ge.Value("basename")
		if e != nil {
			t.Fatalf("failed to get builtin function: %s", e.Error())
		}
		sf, ok := v.(FunctionValue)
		if !ok {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
		if sf.value.name != "basename" {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
	}

	{
		v, e := ge.Value("dirname")
		if e != nil {
			t.Fatalf("failed to get builtin function: %s", e.Error())
		}
		sf, ok := v.(FunctionValue)
		if !ok {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
		if sf.value.name != "dirname" {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
	}

	{
		v, e := ge.Value("prefix")
		if e != nil {
			t.Fatalf("failed to get builtin function: %s", e.Error())
		}
		sf, ok := v.(FunctionValue)
		if !ok {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
		if sf.value.name != "prefix" {
			t.Fatalf("failed to get builtin function: %s", sf)
		}
	}

	if e := ge.Assign("hoge", StringValue{"foo"}); e != nil {
		t.Fatalf("Cannot assign value: %s", e)
	}
	if e := ge.Assign("foo", IntValue{123}); e != nil {
		t.Fatalf("Cannot assign value: %s", e)
	}

	tempArray := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}}
	if e := ge.Assign("bar", ArrayValue{tempArray[:]}); e != nil {
		t.Fatalf("Cannot assign value: %s", e)
	}

	if v, e := ge.Value("hoge"); e != nil || v != (StringValue{"foo"}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := ge.Value("foo"); e != nil || v != (IntValue{123}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	//if v, e := ge.Value("bar"); e != nil || v != (ArrayValue{tempArray[:]}) {
	//	t.Fatalf("Invalid value: %s / error: %e", v, e)
	//}
	if v, e := ge.Value("xxx"); e.Error() != "Unknown variable xxx" || v != nil {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}

	se := CreateSubEnvironment(ge)
	if se.ParentEnvironment() != ge {
		t.Fatalf("Invalid parent environment: %p", se.ParentEnvironment())
	}
	if e := se.Assign("hoge", StringValue{"x"}); e != nil {
		t.Fatalf("Cannot assign value: %s", e)
	}
	if e := se.Assign("foo2", IntValue{456}); e != nil {
		t.Fatalf("Cannot assign value: %s", e)
	}

	if v, e := se.Value("hoge"); e != nil || v != (StringValue{"x"}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := se.Value("foo"); e != nil || v != (IntValue{123}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := se.Value("foo2"); e != nil || v != (IntValue{456}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	//if v, e := se.Value("bar"); e != nil || v != (ArrayValue{[...]int{1, 2, 3}}) {
	//	t.Fatalf("Invalid value: %s / error: %e", v, e)
	//}
	if v, e := se.Value("xxx"); e.Error() != "Unknown variable xxx" || v != nil {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}

	nm := make(map[string]Value)
	nm["hoge"] = StringValue{"y"}
	nm["123"] = IntValue{0}
	se2 := CreateMixedEnvironment(se, nm)

	if v, e := se2.Value("hoge"); e != nil || v != (StringValue{"y"}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := se2.Value("foo"); e != nil || v != (IntValue{123}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := se2.Value("foo2"); e != nil || v != (IntValue{456}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	//if v, e := se2.Value("bar"); e != nil || v != (Value{[...]int{1, 2, 3}}) {
	//	t.Fatalf("Invalid value: %s / error: %e", v, e)
	//}
	if v, e := se2.Value("123"); e != nil || v != (IntValue{0}) {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
	if v, e := se2.Value("xxx"); e.Error() != "Unknown variable xxx" || v != nil {
		t.Fatalf("Invalid value: %s / error: %e", v, e)
	}
}
