package flowscript

import (
	"encoding/json"
	"sort"
)

type StringSet struct {
	set map[string]struct{}
}

func NewStringSet() StringSet {
	return StringSet{
		set: make(map[string]struct{}),
	}
}

func NewStringSetWithValues(values ...string) StringSet {
	m := make(map[string]struct{})

	for _, v := range values {
		m[v] = struct{}{}
	}

	return StringSet{set: m}
}

func (v StringSet) Add(s string) {
	v.set[s] = struct{}{}
}

func (v StringSet) AddAll(s StringSet) {
	for k, _ := range s.set {
		v.Add(k)
	}
}

func (v StringSet) Remove(s string) {
	delete(v.set, s)
}

func (v StringSet) Contains(s string) bool {
	_, ok := v.set[s]
	return ok
}

func (v StringSet) Size() int {
	return len(v.set)
}

func (v StringSet) Array() []string {
	values := make([]string, len(v.set))
	i := 0
	for k, _ := range v.set {
		values[i] = k
		i += 1
	}
	sort.Strings(values)
	return values
}

func (v StringSet) Intersect(x StringSet) StringSet {
	n := NewStringSet()
	for k, _ := range x.set {
		if v.Contains(k) {
			n.Add(k)
		}
	}
	return n
}

func (v StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Array())
}

func (v *StringSet) UnmarshalJSON(data []byte) error {
	var items []string
	err := json.Unmarshal(data, &items)
	if err != nil {
		return err
	}

	v.set = make(map[string]struct{})
	for _, x := range items {
		v.set[x] = struct{}{}
	}
	return nil
}
