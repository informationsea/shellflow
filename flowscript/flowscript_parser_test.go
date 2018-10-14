package flowscript

import (
	"reflect"
	"strings"
	"testing"
)

func createInitializedTokenizer(text string) *LookAheadScanner {
	scanner := NewTokenizerFromText(text)
	if !scanner.Scan() {
		panic("failed to scan")
	}
	return scanner
}

func checkNumberLevel(t *testing.T, p parser) {
	if v, e := p(createInitializedTokenizer("123")); e != nil || v != (ValueEvaluable{IntValue{123}}) {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}

	{
		tokenier := createInitializedTokenizer("123 456")
		if v, e := p(tokenier); e != nil || v != (ValueEvaluable{IntValue{123}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		if v, e := p(tokenier); e != nil || v != (ValueEvaluable{IntValue{456}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
	}

	if v, e := p(createInitializedTokenizer("2147483647")); e != nil || v != (ValueEvaluable{IntValue{2147483647}}) {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}

	// out of range
	if v, e := p(createInitializedTokenizer("2147483648")); e == nil || v != nil {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func TestParseAsNumber(t *testing.T) {
	checkNumberLevel(t, ParseAsNumber)

	// unmatched
	if v, e := ParseAsNumber(createInitializedTokenizer("hoge")); e != errUnmatched || v != nil {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func checkStringLevel(t *testing.T, p parser) {
	if v, e := p(createInitializedTokenizer("\"hoge\"")); e != nil || v != (ValueEvaluable{StringValue{"hoge"}}) {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}

	if v, e := p(createInitializedTokenizer("\"foo\\r\\n\\thoge\"")); e != nil || v != (ValueEvaluable{StringValue{"foo\r\n\thoge"}}) {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}

	// bad escape
	if v, e := p(createInitializedTokenizer("\"ho\\xge\"")); e == nil || v != nil {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func TestParseAsString(t *testing.T) {
	checkStringLevel(t, ParseAsString)

	// unmatched
	if v, e := ParseAsString(createInitializedTokenizer("foo")); e != errUnmatched || v != nil {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func TestParseAsVariable(t *testing.T) {
	if v, e := ParseAsVariable(createInitializedTokenizer("hoge")); e != nil {
		v1, ok := v.(*Variable)
		if !ok || v1.Name != "hoge" {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
	}

	if v, e := ParseAsVariable(createInitializedTokenizer("12foo")); e == nil || v != nil || e != errUnmatched {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}

	if v, e := ParseAsVariable(createInitializedTokenizer("[")); e == nil || v != nil || e != errUnmatched {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func checkFactor3Level(t *testing.T, p parser) {
	checkNumberLevel(t, p)
	checkStringLevel(t, p)
	checkArrayAccessLevel(t, p)
	checkFunctionCallParseLevel(t, p)
	ge := createTestGlobalEnvironment()
	{
		v, e := p(createInitializedTokenizer("hoge"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		v1, ok := v.(*Variable)
		if !ok || v1.Name != "hoge" {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (StringValue{"hoge"}) {
			t.Fatalf("Bad evaluated result: %s", ev)
		}
	}

	{
		v, e := p(createInitializedTokenizer("\"hoge\""))
		if e != nil || v != (ValueEvaluable{StringValue{"hoge"}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (StringValue{"hoge"}) {
			t.Fatalf("Bad evaluated result: %s", ev)
		}
	}

	{
		v, e := p(createInitializedTokenizer("123"))
		if e != nil || v != (ValueEvaluable{IntValue{123}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{123}) {
			t.Fatalf("Bad evaluated result: %s", ev)
		}
	}
}

func TestParseAsFactor3(t *testing.T) {
	checkFactor3Level(t, ParseAsFactor3)

	if v, e := ParseAsFactor3(createInitializedTokenizer("<")); e == nil || v != nil || e != errUnmatched {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func checkFactor2Level(t *testing.T, p parser) {
	checkFactor3Level(t, p)
	ge := createTestGlobalEnvironment()

	{
		v, e := p(createInitializedTokenizer("4 * 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*NumericOperationExpression)
		if !ok || v1.exp1 != (ValueEvaluable{IntValue{4}}) || v1.exp2 != (ValueEvaluable{IntValue{2}}) || v1.operator != "*" {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (IntValue{8}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("4 / 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*NumericOperationExpression)
		if !ok || v1.exp1 != (ValueEvaluable{IntValue{4}}) || v1.exp2 != (ValueEvaluable{IntValue{2}}) || v1.operator != "/" {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (IntValue{2}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}

	if v, e := p(createInitializedTokenizer("4 / !")); e == nil || e.Error() != "parse error: 4 / !" {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func TestParseAsFactor2(t *testing.T) {
	checkFactor2Level(t, ParseAsFactor2)

	if v, e := ParseAsFactor2(createInitializedTokenizer("<")); e == nil || v != nil || e != errUnmatched {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func checkFactor1Level(t *testing.T, p parser) {
	checkFactor2Level(t, p)
	ge := createTestGlobalEnvironment()

	{
		v, e := p(createInitializedTokenizer("4 - 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*NumericOperationExpression)
		if !ok || v1.exp1 != (ValueEvaluable{IntValue{4}}) || v1.exp2 != (ValueEvaluable{IntValue{2}}) || v1.operator != "-" {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (IntValue{2}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("4 + 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*PlusExpression)
		if !ok || v1.exp1 != (ValueEvaluable{IntValue{4}}) || v1.exp2 != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (IntValue{6}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("4 + \"hoge\""))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*PlusExpression)
		if !ok || v1.exp1 != (ValueEvaluable{IntValue{4}}) || v1.exp2 != (ValueEvaluable{StringValue{"hoge"}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (StringValue{"4hoge"}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("\"hoge\" + 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		v1, ok := v.(*PlusExpression)
		if !ok || v1.exp1 != (ValueEvaluable{StringValue{"hoge"}}) || v1.exp2 != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		ev, ee := v.Evaluate(ge)
		if ee != nil || ev != (StringValue{"hoge2"}) {
			t.Fatalf("Bad evaluate result %d / error: %s", ev, ee)
		}
	}
}

func TestParseAsFactor1(t *testing.T) {
	checkFactor1Level(t, ParseAsFactor1)
	if v, e := ParseAsFactor1(createInitializedTokenizer("<")); e == nil || v != nil || e != errUnmatched {
		t.Fatalf("Bad result: %s / error: %s", v, e)
	}
}

func checkFactor0Level(t *testing.T, p parser) {
	checkFactor1Level(t, p)

	ge := createTestGlobalEnvironment()

	{
		if gv, gerr := ge.Value("hoge"); gerr != nil || gv != (StringValue{"hoge"}) {
			t.Fatalf("Bad assigned value %s / error: %s", gv, gerr)
		}

		v, e := p(createInitializedTokenizer("hoge = 2"))
		if e != nil {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		v1, ok := v.(*AssignExpression)
		if !ok {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		left, ok := v1.variable.(*Variable)
		if !ok {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}
		if left.Name != "hoge" || v1.exp != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("Bad result: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ev != (IntValue{2}) {
			t.Fatalf("Bad result: %s / error: %s", ev, ee)
		}

		if gv, gerr := ge.Value("hoge"); gerr != nil || gv != (IntValue{2}) {
			t.Fatalf("Bad assigned value %s / error: %s", gv, gerr)
		}
	}
}

func TestParseAsFactor0(t *testing.T) {
	checkFactor0Level(t, ParseAsFactor0)
}

func checkExpLevel(t *testing.T, p parser) {
	checkFactor0Level(t, p)

	ge := createTestGlobalEnvironment()

	{
		v, e := p(createInitializedTokenizer("1; 3"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj, ok := v.(*JoinedExpression)
		if !ok || vj.exp1 != (ValueEvaluable{IntValue{1}}) || vj.exp2 != (ValueEvaluable{IntValue{3}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{3}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("1; 3; 4"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj, ok := v.(*JoinedExpression)
		if !ok || vj.exp1 != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}
		vj2, ok2 := vj.exp2.(*JoinedExpression)
		if !ok2 || vj2.exp1 != (ValueEvaluable{IntValue{3}}) || vj2.exp2 != (ValueEvaluable{IntValue{4}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{4}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}
	}
}

func TestParseAsExp(t *testing.T) {
	checkExpLevel(t, ParseAsExp)

	ge := createTestGlobalEnvironment()

	{
		v, e := ParseAsExp(createInitializedTokenizer("1 + 3 * 2"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj1, ok := v.(*PlusExpression)
		if !ok || vj1.exp1 != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj2, ok := vj1.exp2.(*NumericOperationExpression)
		if !ok || vj2.exp1 != (ValueEvaluable{IntValue{3}}) || vj2.exp2 != (ValueEvaluable{IntValue{2}}) || vj2.operator != "*" {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{7}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}

	}

	{
		v, e := ParseAsExp(createInitializedTokenizer("1 + (3 * 2)"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj1, ok := v.(*PlusExpression)
		if !ok || vj1.exp1 != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj2, ok := vj1.exp2.(*NumericOperationExpression)
		if !ok || vj2.exp1 != (ValueEvaluable{IntValue{3}}) || vj2.exp2 != (ValueEvaluable{IntValue{2}}) || vj2.operator != "*" {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{7}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}

	}

	{
		v, e := ParseAsExp(createInitializedTokenizer("(1 + 3) * 2"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj1, ok := v.(*NumericOperationExpression)

		if !ok || vj1.exp2 != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj2, ok := vj1.exp1.(*PlusExpression)
		if !ok || vj2.exp1 != (ValueEvaluable{IntValue{1}}) || vj2.exp2 != (ValueEvaluable{IntValue{3}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{8}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}

	}

	{
		v, e := ParseAsExp(createInitializedTokenizer("1 + 3 + 2"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj1, ok := v.(*PlusExpression)

		if !ok || vj1.exp1 != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj2, ok := vj1.exp2.(*PlusExpression)
		if !ok || vj2.exp1 != (ValueEvaluable{IntValue{3}}) || vj2.exp2 != (ValueEvaluable{IntValue{2}}) {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{6}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}

	}

	{
		v, e := ParseAsExp(createInitializedTokenizer("4 * 8 / 2"))
		if e != nil {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj1, ok := v.(*NumericOperationExpression)

		if !ok || vj1.exp1 != (ValueEvaluable{IntValue{4}}) || vj1.operator != "*" {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		vj2, ok := vj1.exp2.(*NumericOperationExpression)
		if !ok || vj2.exp1 != (ValueEvaluable{IntValue{8}}) || vj2.exp2 != (ValueEvaluable{IntValue{2}}) || vj2.operator != "/" {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{16}) {
			t.Fatalf("Bad evaluated value: %s / error: %s", ev, ee)
		}

	}

	{
		v, e := ParseAsExp(createInitializedTokenizer("(1"))
		if e == nil || v != nil || !strings.HasPrefix(e.Error(), "syntax error: ) is not found") {
			t.Fatalf("Failed to parse: %s / error: %s", v, e)
		}
	}
}

func checkArrayAccessLevel(t *testing.T, p parser) {
	checkArrayLevel(t, p)

	ge := createTestGlobalEnvironment()
	{
		tokenizer := createInitializedTokenizer("foo[hoge];")
		v, e := p(tokenizer)
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		if tokenizer.Text() != ";" {
			t.Fatalf("bad tokenizer position: %s", tokenizer.Text())
		}

		arrayAccess, ok := v.(*ArrayAccess)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if array, ok := arrayAccess.Array.(*Variable); !ok || array.Name != "foo" {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayIndex, ok := arrayAccess.ArrayIndex.(*Variable); !ok || arrayIndex.Name != "hoge" {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[1,2,3,4][2]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess, ok := v.(*ArrayAccess)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		// TODO: fix here
		expected := [...]Value{IntValue{1}, IntValue{2}, IntValue{3}, IntValue{4}}

		if array, ok := arrayAccess.Array.(*ArrayExpression); !ok || reflect.DeepEqual(array.values, expected) {
			t.Fatalf("Bad parse result: %s", v)
		}

		arrayIndexEvauable, ok := arrayAccess.ArrayIndex.(ValueEvaluable)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayIndex, ok := arrayIndexEvauable.value.(IntValue); !ok || arrayIndex != (IntValue{2}) {
			t.Fatalf("Bad parse result: %s", v)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (IntValue{3}) {
			t.Fatalf("Bad evaluation result: %s / error: %s", ev, ee)
		}
	}

	{
		v, e := p(createInitializedTokenizer("foo[10][20]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess1, ok := v.(*ArrayAccess)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		arrayAccess2, ok := arrayAccess1.Array.(*ArrayAccess)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if array, ok := arrayAccess2.Array.(*Variable); !ok || array.Name != "foo" {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayIndex, ok := arrayAccess2.ArrayIndex.(ValueEvaluable); !ok || arrayIndex.value != (IntValue{10}) {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayIndex, ok := arrayAccess1.ArrayIndex.(ValueEvaluable); !ok || arrayIndex.value != (IntValue{20}) {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess1, ok := v.(*ArrayExpression)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}
		if len(arrayAccess1.values) != 0 {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[ 1 ,   2 ,3  ] "))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess1, ok := v.(*ArrayExpression)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}
		if len(arrayAccess1.values) != 3 {
			t.Fatalf("Bad parse result: %s", v)
		}
		if av, ok := arrayAccess1.values[0].(ValueEvaluable); !ok || av.value != (IntValue{1}) {
			t.Fatalf("Bad parse result: %s", v)
		}
		if av, ok := arrayAccess1.values[1].(ValueEvaluable); !ok || av.value != (IntValue{2}) {
			t.Fatalf("Bad parse result: %s", v)
		}
		if av, ok := arrayAccess1.values[2].(ValueEvaluable); !ok || av.value != (IntValue{3}) {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("foo[hoge)"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := p(createInitializedTokenizer("foo[(1])"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := p(createInitializedTokenizer("foo[!)"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}
}

func TestParseAsArrayAccess(t *testing.T) {
	checkArrayAccessLevel(t, ParseAsArrayAccessOrArray)

	{
		v, e := ParseAsArrayAccessOrArray(createInitializedTokenizer("10"))
		if e != errUnmatched {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := ParseAsArrayAccessOrArray(createInitializedTokenizer("hoge + 10"))
		if e != errUnmatched {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}
}

func checkArrayLevel(t *testing.T, p parser) {
	{
		v, e := p(createInitializedTokenizer("[1,hoge,3]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess, ok := v.(*ArrayExpression)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if len(arrayAccess.values) != 3 {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayAccess.values[0] != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Bad parse result: %s", v)
		}

		if v, ok := arrayAccess.values[1].(*Variable); !ok || v.Name != "hoge" {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayAccess.values[2] != (ValueEvaluable{IntValue{3}}) {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[1]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess, ok := v.(*ArrayExpression)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if len(arrayAccess.values) != 1 {
			t.Fatalf("Bad parse result: %s", v)
		}

		if arrayAccess.values[0] != (ValueEvaluable{IntValue{1}}) {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[]"))
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		arrayAccess, ok := v.(*ArrayExpression)
		if !ok {
			t.Fatalf("Bad parse result: %s", v)
		}

		if len(arrayAccess.values) != 0 {
			t.Fatalf("Bad parse result: %s", v)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[1)"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := p(createInitializedTokenizer("["))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[(1])"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		v, e := p(createInitializedTokenizer("[#]"))
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error: no expression is found:") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}
}

func TestParseAsArray(t *testing.T) {
	checkArrayLevel(t, ParseAsArray)
}

func checkFunctionCallParseLevel(t *testing.T, p parser) {
	ge := createTestGlobalEnvironment()
	{
		tokenizer := createInitializedTokenizer("basename(\"hoge/foo\");")
		v, e := p(tokenizer)
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		if tokenizer.Text() != ";" {
			t.Fatalf("Wrong tokenizer position: %s", tokenizer.Text())
		}
		fv, ok := v.(*FunctionCall)
		if !ok {
			t.Fatalf("Failed to parse %s", v)
		}

		vv, ok := fv.function.(*Variable)
		if !ok || vv.Name != "basename" {
			t.Fatalf("Failed to parse %s", fv)
		}

		if len(fv.args) != 1 || fv.args[0] != (ValueEvaluable{StringValue{"hoge/foo"}}) {
			t.Fatalf("Failed to parse %s", fv.args[0])
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (StringValue{"foo"}) {
			t.Fatalf("Bad evaluate result: %s / error: %s", ev, ee)
		}
	}

	{
		tokenizer := createInitializedTokenizer("basename();")
		v, e := p(tokenizer)
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		if tokenizer.Text() != ";" {
			t.Fatalf("Wrong tokenizer position: %s", tokenizer.Text())
		}
		fv, ok := v.(*FunctionCall)
		if !ok {
			t.Fatalf("Failed to parse %s", v)
		}

		vv, ok := fv.function.(*Variable)
		if !ok || vv.Name != "basename" {
			t.Fatalf("Failed to parse %s", fv)
		}

		if len(fv.args) != 0 {
			t.Fatalf("Failed to parse %s", fv)
		}
	}

	{
		tokenizer := createInitializedTokenizer("basename(\"hoge/foo.c\", \".c\");")
		v, e := p(tokenizer)
		if e != nil {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		if tokenizer.Text() != ";" {
			t.Fatalf("Wrong tokenizer position: %s", tokenizer.Text())
		}
		fv, ok := v.(*FunctionCall)
		if !ok {
			t.Fatalf("Failed to parse %s", v)
		}

		vv, ok := fv.function.(*Variable)
		if !ok || vv.Name != "basename" {
			t.Fatalf("Failed to parse %s", fv)
		}

		if len(fv.args) != 2 || fv.args[0] != (ValueEvaluable{StringValue{"hoge/foo.c"}}) || fv.args[1] != (ValueEvaluable{StringValue{".c"}}) {
			t.Fatalf("Failed to parse %s", fv)
		}

		if ev, ee := v.Evaluate(ge); ee != nil || ev != (StringValue{"foo"}) {
			t.Fatalf("Bad evaluate result: %s / error: %s", ev, ee)
		}
	}

	{
		tokenizer := createInitializedTokenizer("basename(\"hoge/foo.c\", \".c\"]")
		v, e := p(tokenizer)
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		tokenizer := createInitializedTokenizer("basename(]")
		v, e := p(tokenizer)
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}

	{
		tokenizer := createInitializedTokenizer("basename((12])")
		v, e := p(tokenizer)
		if e == nil || !strings.HasPrefix(e.Error(), "syntax error") {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
	}
}

func TestFunctionCallParse(t *testing.T) {
	checkFunctionCallParseLevel(t, ParseAsFunctionCall)

	{
		tokenizer := createInitializedTokenizer("123(\"hoge/foo.c\", \".c\");")
		v, e := ParseAsFunctionCall(tokenizer)
		if e != errUnmatched {
			t.Fatalf("Failed to parse %s / error: %s", v, e)
		}
		if tokenizer.Text() != "123" {
			t.Fatalf("Wrong tokenizer position: %s", tokenizer.Text())
		}
	}
}
