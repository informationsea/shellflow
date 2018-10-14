package flowscript

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsCharacter(t *testing.T) {
	for _, v := range " \n\r\f\t" {
		if r := IsWhiteSpace(rune(v)); !r {
			t.Fatalf("Bad white space check result: %#U", v)
		}
	}

	for _, v := range "0123456789" {
		if r := IsDigit(rune(v)); !r {
			t.Fatalf("Bad digit check result: %#U", v)
		}
	}

	for _, v := range "0123456789_abcxyzABCXYZ" {
		if r := IsWordCharacter(rune(v)); !r {
			t.Fatalf("Bad word character check result: %#U", v)
		}
	}
}

func TestTakeWhile(t *testing.T) {
	if l := takeRuneWhile([]byte(" \nx"), IsWhiteSpace); l != 2 {
		t.Fatalf("Bad takeRuneWhile result: %d", l)
	}

	if l := takeRuneWhile([]byte("012x"), IsDigit); l != 3 {
		t.Fatalf("Bad takeRuneWhile result: %d", l)
	}

	if l := takeRuneWhile([]byte("012x"), IsWhiteSpace); l != 0 {
		t.Fatalf("Bad takeRuneWhile result: %d", l)
	}

	if l := takeRuneWhile([]byte(""), IsWhiteSpace); l != 0 {
		t.Fatalf("Bad takeRuneWhile result: %d", l)
	}

	if l := takeRuneWhile([]byte("   "), IsWhiteSpace); l != 3 {
		t.Fatalf("Bad takeRuneWhile result: %d", l)
	}
}

func checkSplitResult(t *testing.T, data string, expected string, expectedAdvance int, atEOF bool) {
	advance, token, err := SplitToken([]byte(data), atEOF)

	if err != nil {
		t.Fatalf("testing data: %s  / error %s", data, err)
	}

	if advance != expectedAdvance {
		t.Fatalf("invalid advance length %d / expected %d : %s", advance, expectedAdvance, token)
	}

	if expected == "" {
		if token != nil {
			t.Fatalf("Wrong token %s(%d)", token, len(token))
		}
	} else {
		if !reflect.DeepEqual(token, []byte(expected)) {
			t.Fatalf("Wrong token %s(%d) / expected %s(%d)", token, len(token), expected, len(expected))
		}
	}
}

func TestSplitFunc(t *testing.T) {
	checkSplitResult(t, "hello,", "hello", 5, false)
	checkSplitResult(t, "  hello,", "hello", 7, false)
	checkSplitResult(t, "  ,, false)", ",", 3, false)
	checkSplitResult(t, "\n (hoge, false)", "(", 3, false)
	checkSplitResult(t, "123a", "123", 3, false)
	checkSplitResult(t, "foo123;", "foo123", 6, false)
	checkSplitResult(t, ";3", ";", 1, false)

	checkSplitResult(t, "123", "", 0, false)
	checkSplitResult(t, "abc123", "", 0, false)
	checkSplitResult(t, "  abc123", "", 2, false)
	checkSplitResult(t, "\"abc", "", 0, false)

	checkSplitResult(t, "\"foo bar\" 123", "\"foo bar\"", 9, false)
	checkSplitResult(t, "\"foo\\\" bar\" 123", "\"foo\\\" bar\"", 11, false)

	checkSplitResult(t, "123", "123", 3, true)
	checkSplitResult(t, "abc123", "abc123", 6, true)
	checkSplitResult(t, "  abc123", "abc123", 8, true)
	checkSplitResult(t, "\"abc", "\"abc", 4, true)

}

func TestScanner(t *testing.T) {
	expectedTokens := [...]string{"hello", "=", "\"123\"", ";", "[", "1", ",", "2", ",", "3", "]", ";", "foo", "(", "hoge", ")"}

	reader := strings.NewReader("hello = \"123\"; [1,2,3]; foo(hoge)")
	scanner := NewTokenizer(reader)

	for _, v := range expectedTokens {
		if s := scanner.Scan(); !s {
			t.Fatal("Failed to scan")
		}
		if x := scanner.Text(); x != v {
			t.Fatalf("Bad token: %s / expected: %s", x, v)
		}
	}

	if scanner.Scan() {
		t.Fatalf("Should reached to EOF: %s (%d)", scanner.Bytes(), len(scanner.Bytes()))
	}

	if scanner.Scan() {
		t.Fatalf("Should reached to EOF: %s (%d)", scanner.Bytes(), len(scanner.Bytes()))
	}

	if v := scanner.LookAheadBytes(1); v != nil {
		t.Fatalf("Should reached to EOF: %s (%d)", v, len(v))
	}
}

func TestScanner2(t *testing.T) {
	expectedTokens := [...]string{"hello", "=", "1"}

	reader := strings.NewReader("hello = 1 ")
	scanner := NewTokenizer(reader)

	for i, v := range expectedTokens {
		if x := scanner.LookAheadText(i); x != v {
			t.Fatalf("Bad token: %s / expected: %s", x, v)
		}
	}

	for _, v := range expectedTokens {
		if s := scanner.Scan(); !s {
			t.Fatal("Failed to scan")
		}
		if x := scanner.Text(); x != v {
			t.Fatalf("Bad token: %s / expected: %s", x, v)
		}
	}

	if scanner.Scan() {
		t.Fatalf("Should reached to EOF: %s (%d)", scanner.Bytes(), len(scanner.Bytes()))
	}
}
