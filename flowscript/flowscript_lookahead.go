package flowscript

import (
	"bufio"
	"container/list"
)

type LookAheadScanner struct {
	scanner      *bufio.Scanner
	lookAhead    *list.List
	lastestError error
	firstScan    bool
}

func NewLookAheadScanner(scanner *bufio.Scanner) *LookAheadScanner {
	return &LookAheadScanner{
		scanner:      scanner,
		lookAhead:    list.New(),
		lastestError: nil,
		firstScan:    true,
	}
}

func (s *LookAheadScanner) Bytes() []byte {
	if s.lookAhead.Len() != 0 {
		if v, ok := s.lookAhead.Front().Value.([]byte); ok {
			return v
		}
		panic("Bad type")
	}
	return nil
}

func (s *LookAheadScanner) Text() string {
	return string(s.Bytes())
}

func (s *LookAheadScanner) Scan() bool {
	if s.lastestError != nil {
		return false
	}

	if s.lookAhead.Len() > 0 && !s.firstScan {
		s.lookAhead.Remove(s.lookAhead.Front())
	}

	s.firstScan = false

	if s.lookAhead.Len() == 0 {
		if !s.scanner.Scan() {
			s.lastestError = s.Err()
			return false
		}
		s.lookAhead.PushBack(s.scanner.Bytes())
	}
	return true
}

func (s *LookAheadScanner) LookAheadBytes(i int) []byte {
	if s.lastestError != nil {
		return nil
	}

	for s.lookAhead.Len() < (i + 1) {
		if !s.scanner.Scan() {
			s.lastestError = s.Err()
			return nil
		}
		s.lookAhead.PushBack(s.scanner.Bytes())
	}

	v := s.lookAhead.Front()
	for j := 0; j < i; j++ {
		v = v.Next()
	}

	if b, ok := v.Value.([]byte); ok {
		return b
	}

	panic("bad type")
}

func (s *LookAheadScanner) LookAheadText(i int) string {
	return string(s.LookAheadBytes(i))
}

func (s *LookAheadScanner) Err() error {
	return s.lastestError
}
