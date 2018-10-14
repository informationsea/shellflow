package flowscript

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

/*
 * Syntax
 * exp := <factor0> ; <factor0> | <factor0>
 * factor0 := <factor1> = <factor1> | <factor1>
 * factor1 := <factor2> + <factor1> | <factor2> - <factor1> | <factor2>
 * factor2 := <factor3> * <factor2> | <factor3> / <factor2> | <factor3>
 * factor3 := <array_access> | <function_call> | <string> | <number> | <ident> | ( <exp> )
 * array_access_or_array := <ident> [ <exp> ] | <array_access> [ <exp> ] | <array> [ <exp> ] | <array>
 * array := [ <exp> {, <exp>}* ]
 * function_call := <ident> ( {<exp> {, <exp>}*}? )
 */

var numberRegexp = regexp.MustCompile("\\d+")
var variableRegexp = regexp.MustCompile("^[a-zA-Z_]\\w*$")
var errUnmatched = errors.New("Unmatched")
var errFinished = errors.New("Finished")

type isMatched func(token []byte) bool
type convertToType func(token []byte) (eval Evaluable, err error)
type parser func(tokenizer *LookAheadScanner) (eval Evaluable, err error)
type binaryEvaluableCreator func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error)

func parserHelper(tokenizer *LookAheadScanner, test isMatched, convert convertToType) (eval Evaluable, err error) {
	token := tokenizer.Bytes()
	if test(token) {
		r, e := convert(token)
		if e != nil {
			return nil, e
		}
		tokenizer.Scan()
		if te := tokenizer.Err(); te != nil {
			return nil, te
		} else {
			return r, nil
		}
	}
	return nil, errUnmatched
}

func binaryOperatorParserHelper(tokenizer *LookAheadScanner, leftParser parser, rightParser parser, binary map[string]binaryEvaluableCreator) (eval Evaluable, err error) {
	eval1, err := leftParser(tokenizer)
	if err == errUnmatched {
		return nil, errUnmatched
	}
	if err == nil {
		next := tokenizer.Text()
		creator, ok := binary[next]
		if ok && len(tokenizer.LookAheadBytes(1)) > 0 {
			tokenizer.Scan()
			if te := tokenizer.Err(); te != nil {
				return nil, te
			}
			eval2, err2 := rightParser(tokenizer)
			if err2 == nil {
				return creator(eval1, next, eval2)
			} else {
				return nil, fmt.Errorf("parse error: %s %s %s", eval1, next, tokenizer.Text())
			}
		}
		return eval1, nil
	}
	return nil, err
}

func ParseAsNumber(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	if len(tokenizer.Bytes()) < 1 {
		return nil, errUnmatched
	}

	return parserHelper(tokenizer, func(token []byte) bool {
		return numberRegexp.Match(token)
	}, func(token []byte) (eval Evaluable, err error) {
		n, err := strconv.ParseInt(string(token), 10, 32)
		if err == nil {
			eval = ValueEvaluable{IntValue{n}}
		}
		return
	})
}

func ParseAsString(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	if len(tokenizer.Bytes()) < 2 {
		return nil, errUnmatched
	}

	return parserHelper(tokenizer, func(token []byte) bool {
		return token[0] == '"' && token[len(token)-1] == '"'
	}, func(token []byte) (eval Evaluable, err error) {
		n, err := strconv.Unquote(string(token))
		if err == nil {
			eval = ValueEvaluable{StringValue{n}}
		}
		return
	})
}

func ParseAsVariable(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	return parserHelper(tokenizer, func(token []byte) bool {
		return variableRegexp.Match(token)
	}, func(token []byte) (eval Evaluable, err error) {
		return &Variable{string(token)}, nil
	})
}

var parseAsExpMap = map[string]binaryEvaluableCreator{
	";": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &JoinedExpression{exp1: exp1, exp2: exp2}, nil
	},
}

func ParseAsExp(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	return binaryOperatorParserHelper(tokenizer, ParseAsFactor0, ParseAsExp, parseAsExpMap)
}

var parseAsFactor0Map = map[string]binaryEvaluableCreator{
	"=": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &AssignExpression{variable: exp1, exp: exp2}, nil
	},
}

func ParseAsFactor0(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	return binaryOperatorParserHelper(tokenizer, ParseAsFactor1, ParseAsFactor1, parseAsFactor0Map)
}

var parseAsFactor1Map = map[string]binaryEvaluableCreator{
	"+": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &PlusExpression{exp1: exp1, exp2: exp2}, nil
	},
	"-": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &NumericOperationExpression{exp1: exp1, exp2: exp2, operator: operator}, nil
	},
}

func ParseAsFactor1(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	return binaryOperatorParserHelper(tokenizer, ParseAsFactor2, ParseAsFactor1, parseAsFactor1Map)
}

var parseAsFactor2Map = map[string]binaryEvaluableCreator{
	"*": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &NumericOperationExpression{exp1: exp1, exp2: exp2, operator: operator}, nil
	},
	"/": func(exp1 Evaluable, operator string, exp2 Evaluable) (eval Evaluable, err error) {
		return &NumericOperationExpression{exp1: exp1, exp2: exp2, operator: operator}, nil
	},
}

func ParseAsFactor2(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	return binaryOperatorParserHelper(tokenizer, ParseAsFactor3, ParseAsFactor2, parseAsFactor2Map)
}

func ParseAsFactor3(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	token := tokenizer.Text()
	if token == "(" {
		tokenizer.Scan()
		if e := tokenizer.Err(); e != nil {
			return nil, e
		}

		eval, err = ParseAsExp(tokenizer)

		endToken := tokenizer.Text()
		if endToken != ")" {
			return nil, fmt.Errorf("syntax error: ) is not found: ( %s %s", eval.String(), endToken)
		}

		tokenizer.Scan()
		if e := tokenizer.Err(); e != nil {
			return nil, e
		}
		return
	}

	eval, err = ParseAsArrayAccessOrArray(tokenizer)
	if err != errUnmatched {
		return
	}

	eval, err = ParseAsFunctionCall(tokenizer)
	if err != errUnmatched {
		return
	}

	eval, err = ParseAsString(tokenizer)
	if err != errUnmatched {
		return
	}
	eval, err = ParseAsNumber(tokenizer)
	if err != errUnmatched {
		return
	}
	eval, err = ParseAsVariable(tokenizer)
	if err != errUnmatched {
		return
	}

	return nil, errUnmatched
}

func ParseAsArray(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	if tokenizer.Text() != "[" {
		return nil, errUnmatched
	}
	tokenizer.Scan()
	if e := tokenizer.Err(); e != nil {
		return nil, e
	}

	var values []Evaluable
	for {
		if tokenizer.Text() == "]" {
			break
		}
		current, err := ParseAsExp(tokenizer)
		if err == errUnmatched {
			return nil, fmt.Errorf("syntax error: no expression is found: %s", tokenizer.Text())
		}
		if err != nil {
			return nil, err
		}
		values = append(values, current)
		if tokenizer.Text() != "," {
			break
		}
		tokenizer.Scan()
		if e := tokenizer.Err(); e != nil {
			return nil, e
		}
	}

	if tokenizer.Text() != "]" {
		return nil, fmt.Errorf("syntax error: \"]\" is not found: [%s%s ", values, tokenizer.Text())
	}
	tokenizer.Scan()
	if e := tokenizer.Err(); e != nil {
		return nil, e
	}

	return &ArrayExpression{values: values}, nil
}

func ParseAsArrayAccessOrArray(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	var array Evaluable
	err = errUnmatched

	tried := false
	if tokenizer.Text() == "[" {
		array, err = ParseAsArray(tokenizer)
		tried = true
	} else if variableRegexp.Match(tokenizer.Bytes()) && tokenizer.LookAheadText(1) == "[" {
		array, err = ParseAsVariable(tokenizer)
		tried = true
	}

	if err == nil {
		var current Evaluable = array
		for tokenizer.Text() == "[" {
			tokenizer.Scan()
			if e := tokenizer.Err(); e != nil {
				return nil, e
			}
			exp, err2 := ParseAsExp(tokenizer)
			if err2 == errUnmatched {
				return nil, fmt.Errorf("syntax error no expression is found in a bracket: %s[%s", array.String(), tokenizer.Text())
			} else if err2 != nil {
				return nil, err2
			}
			if tokenizer.Text() != "]" {
				return nil, fmt.Errorf("syntax error \"]\" is not found:  %s[%s%s", array.String(), exp.String(), tokenizer.Text())
			}
			tokenizer.Scan()
			if e := tokenizer.Err(); e != nil {
				return nil, e
			}
			current = &ArrayAccess{Array: current, ArrayIndex: exp}
		}

		if v, ok := current.(*ArrayAccess); ok {
			return v, nil
		}
		if v, ok := current.(*ArrayExpression); ok {
			return v, nil
		}
	}

	if tried {
		if err == errUnmatched {
			return nil, fmt.Errorf("syntax error: %s", tokenizer.Text())
		}
		return nil, err
	}
	return nil, errUnmatched
}

func ParseAsFunctionCall(tokenizer *LookAheadScanner) (eval Evaluable, err error) {
	if !variableRegexp.Match(tokenizer.Bytes()) || tokenizer.LookAheadText(1) != "(" {
		return nil, errUnmatched
	}

	function, err := ParseAsVariable(tokenizer)
	tokenizer.Scan() // skip "("
	if e := tokenizer.Err(); e != nil {
		return nil, e
	}

	var values []Evaluable
	for {
		if tokenizer.Text() == ")" {
			break
		}
		current, err := ParseAsExp(tokenizer)
		if err == errUnmatched {
			return nil, fmt.Errorf("syntax error: no expression is found: %s", tokenizer.Text())
		}
		if err != nil {
			return nil, err
		}
		values = append(values, current)
		if tokenizer.Text() != "," {
			break
		}
		tokenizer.Scan()
		if e := tokenizer.Err(); e != nil {
			return nil, e
		}
	}

	if tokenizer.Text() != ")" {
		return nil, fmt.Errorf("syntax error: \")\" is not found: [%s%s ", values, tokenizer.Text())
	}
	tokenizer.Scan()
	if e := tokenizer.Err(); e != nil {
		return nil, e
	}

	return &FunctionCall{function: function, args: values}, nil
}
