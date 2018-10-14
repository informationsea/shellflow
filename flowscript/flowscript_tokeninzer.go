package flowscript

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var whiteSpaceRegexp = regexp.MustCompile("\\s")
var digitRegexp = regexp.MustCompile("\\d")
var wordCharacterRegexp = regexp.MustCompile("\\w")

func checkMatch(ch rune, exp *regexp.Regexp) bool {
	var data [5]byte
	var p = data[:]
	utf8.EncodeRune(p, ch)
	return exp.Match(p)
}

func IsDigit(ch rune) bool {
	return checkMatch(ch, digitRegexp)
}

func IsWhiteSpace(ch rune) bool {
	return checkMatch(ch, whiteSpaceRegexp)
}

func IsWordCharacter(ch rune) bool {
	return checkMatch(ch, wordCharacterRegexp)
}

type RuneCheck func(ch rune) (match bool)

func takeRuneWhile(data []byte, checker RuneCheck) (length int) {
	length = 0
	for {
		ch, l := utf8.DecodeRune(data)
		if checker(ch) {
			length += l
			data = data[l:]
		} else {
			break
		}
	}
	return
}

func SplitToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance = 0
	token = nil
	err = nil

	l := takeRuneWhile(data, IsWhiteSpace)
	advance += l
	data = data[l:]

	firstChar, l3 := utf8.DecodeRune(data)

	if IsDigit(firstChar) {
		l2 := takeRuneWhile(data, IsDigit)
		if len(data) != l2 || atEOF {
			advance += l2
			token = data[:l2]
			err = nil
		}
	} else if IsWordCharacter(firstChar) {
		l2 := takeRuneWhile(data, IsWordCharacter)
		if len(data) != l2 || atEOF {
			advance += l2
			token = data[:l2]
			err = nil
		}
	} else if firstChar == rune('"') {
		escaping := false
		currentLength := l3
		for {
			ch, l2 := utf8.DecodeRune(data[currentLength:])
			currentLength += l2

			if (l2 > 0 && ch == rune('"') && !escaping) || (l2 == 0 && atEOF) {
				advance += currentLength
				token = data[:currentLength]
				err = nil
				return
			} else if l2 == 0 {
				break
			} else if ch == rune('\\') && !escaping {
				escaping = true
			} else {
				escaping = false
			}
		}
	} else if l3 > 0 {
		advance += l3
		token = data[:l3]
		err = nil
	}

	if atEOF && token == nil {
		err = errors.New("FAIL")
	}

	return
}

func NewTokenizer(r io.Reader) *LookAheadScanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitToken)
	return NewLookAheadScanner(scanner)
}

func NewTokenizerFromText(text string) *LookAheadScanner {
	reader := strings.NewReader(text)
	return NewTokenizer(reader)
}
