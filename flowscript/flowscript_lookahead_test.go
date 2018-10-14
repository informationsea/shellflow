package flowscript

import (
	"bufio"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLookAheadScannerFirstHead(t *testing.T) {
	reader := strings.NewReader("1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20")
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	loa := NewLookAheadScanner(scanner)

	for i := 1; i <= 20; i++ {
		index := strconv.FormatInt(int64(i), 10)
		if v := loa.LookAheadBytes(i - 1); !reflect.DeepEqual(v, []byte(index)) {
			t.Fatalf("Bad data: %s / expected: %s", v, index)
		}
	}

	for i := 1; i <= 20; i++ {
		if !loa.Scan() {
			t.Fatalf("Cannot scan %s", loa.Err())
		}

		index := strconv.FormatInt(int64(i), 10)
		if v := loa.Bytes(); !reflect.DeepEqual(v, []byte(index)) {
			t.Fatalf("Bad data: %s / expected %s", v, index)
		}
	}
}

func TestLookAheadScanner(t *testing.T) {
	reader := strings.NewReader("1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20")
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	loa := NewLookAheadScanner(scanner)

	// not scanned yet
	if v := loa.Bytes(); v != nil {
		t.Fatalf("Bad data: %s", v)
	}

	for i := 1; i <= 20; i++ {
		index := strconv.FormatInt(int64(i), 10)
		if !loa.Scan() {
			t.Fatalf("Failed to scan %s", loa.Err())
		}

		if e := loa.Err(); e != nil {
			t.Fatalf("Failed to scan %s", e)
		}

		if v := loa.Text(); v != index {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}

		if v := loa.Bytes(); !reflect.DeepEqual(v, []byte(index)) {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}
	}

	if r := loa.Scan(); r {
		t.Fatal("loa should reached EOF")
	}

	if e := loa.Err(); e != nil {
		t.Fatalf("EOF should not create error: %s", e)
	}
}

func TestLookAheadScanner2(t *testing.T) {
	reader := strings.NewReader("1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20")
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	loa := NewLookAheadScanner(scanner)

	// not scanned yet
	if v := loa.Bytes(); v != nil {
		t.Fatalf("Bad data: %s", v)
	}

	for i := 1; i <= 5; i++ {
		index := strconv.FormatInt(int64(i), 10)
		if !loa.Scan() {
			t.Fatalf("Failed to scan %s", loa.Err())
		}

		if e := loa.Err(); e != nil {
			t.Fatalf("Failed to scan %s", e)
		}

		if v := loa.Text(); v != index {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}

		if v := loa.Bytes(); !reflect.DeepEqual(v, []byte(index)) {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}
	}

	if v := loa.LookAheadBytes(1); !reflect.DeepEqual(v, []byte("6")) {
		t.Fatalf("Bad look ahead: %s", v)
	}

	if loa.lookAhead.Len() != 2 {
		t.Fatalf("Wrong number of look ahead list: %d", loa.lookAhead.Len())
	}

	if v := loa.LookAheadBytes(3); !reflect.DeepEqual(v, []byte("8")) {
		t.Fatalf("Bad look ahead: %s", v)
	}

	if v := loa.LookAheadText(3); v != "8" {
		t.Fatalf("Bad look ahead: %s", v)
	}

	if loa.lookAhead.Len() != 4 {
		t.Fatalf("Wrong number of look ahead list: %d", loa.lookAhead.Len())
	}

	for i := 6; i <= 20; i++ {
		index := strconv.FormatInt(int64(i), 10)
		if !loa.Scan() {
			t.Fatalf("Failed to scan %s", loa.Err())
		}

		if e := loa.Err(); e != nil {
			t.Fatalf("Failed to scan %s", e)
		}

		if v := loa.Text(); v != index {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}

		if v := loa.Bytes(); !reflect.DeepEqual(v, []byte(index)) {
			t.Fatalf("Bad data: %s / length %d", v, len(v))
		}
	}

	if v := loa.LookAheadBytes(1); v != nil {
		t.Fatalf("Bad look ahead: %s", v)
	}

	if r := loa.Scan(); r {
		t.Fatal("loa should reached EOF")
	}

	if e := loa.Err(); e != nil {
		t.Fatalf("EOF should not create error: %s", e)
	}
}
