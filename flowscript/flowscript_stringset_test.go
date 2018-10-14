package flowscript

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStringSet(t *testing.T) {
	stringSet := NewStringSet()
	if s := stringSet.Size(); s != 0 {
		t.Fatalf("size of empty set should 0 / actual: %d", s)
	}

	stringSet.Add("hoge")
	stringSet.Add("hoge")
	stringSet.Add("foo")
	if s := stringSet.Size(); s != 2 {
		t.Fatalf("size of set should 2 / actual: %d", s)
	}

	if !stringSet.Contains("hoge") {
		t.Fatal("hoge should be contained")
	}

	if !stringSet.Contains("foo") {
		t.Fatal("foo should be contained")
	}

	if stringSet.Contains("bar") {
		t.Fatal("bar should not be contained")
	}

	if x := stringSet.Array(); !reflect.DeepEqual(x, []string{"foo", "hoge"}) {
		t.Fatalf("bad array: %s", x)
	}

	if x, e := json.Marshal(stringSet); e != nil || string(x) != "[\"foo\",\"hoge\"]" {
		t.Fatalf("bad json: %s / error: %s", x, e)
	}

	stringSet.Remove("hoge")

	if stringSet.Contains("hoge") {
		t.Fatal("hoge should be contained")
	}

	if !stringSet.Contains("foo") {
		t.Fatal("foo should be contained")
	}

	if stringSet.Contains("bar") {
		t.Fatal("bar should not be contained")
	}

	stringSet2 := NewStringSet()
	stringSet2.Add("1")
	stringSet2.Add("2")
	stringSet2.Add("foo")

	stringSet.AddAll(stringSet2)

	if stringSet.Contains("hoge") {
		t.Fatal("hoge should not be contained")
	}

	if !stringSet.Contains("foo") {
		t.Fatal("foo should be contained")
	}

	if stringSet.Contains("bar") {
		t.Fatal("bar should not be contained")
	}

	if !stringSet.Contains("1") {
		t.Fatal("1 should be contained")
	}

	if !stringSet.Contains("2") {
		t.Fatal("2 should be contained")
	}

	if a := stringSet.Array(); !reflect.DeepEqual([]string{"1", "2", "foo"}, a) {
		t.Fatalf("bad array result: %s", a)
	}

	stringSet3 := NewStringSetWithValues("hoge", "1", "3", "4")
	if s := stringSet3.Size(); s != 4 {
		t.Fatalf("bad size: %d", s)
	}
	if v := stringSet3.Array(); !reflect.DeepEqual(v, []string{"1", "3", "4", "hoge"}) {
		t.Fatalf("bad array: %s", v)
	}
	if v := stringSet3.Intersect(stringSet).Array(); !reflect.DeepEqual(v, []string{"1"}) {
		t.Fatalf("bad array: %s", v)
	}
}

func TestStringSetUnmarshal(t *testing.T) {
	var stringSet StringSet
	if e := json.Unmarshal([]byte("[\"a\", \"b\", \"c\"]"), &stringSet); e != nil {
		t.Fatalf("bad unmarshal: %s", e)
	}
	if v := stringSet.Array(); !reflect.DeepEqual(v, []string{"a", "b", "c"}) {
		t.Fatalf("bad unmarshal result: %s", v)
	}
}
