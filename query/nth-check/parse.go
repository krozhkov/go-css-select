package nthcheck

import (
	"errors"
	"fmt"
	"strings"
)

// Following http://www.w3.org/TR/css3-selectors/#nth-child-pseudo

// Whitespace as per https://www.w3.org/TR/selectors-3/#lex is " \t\r\n\f"
func isWhitespace(r byte) bool {
	return r == 9 || r == 10 || r == 12 || r == 13 || r == 32
}

const ZERO = '0'
const NINE = '9'

/**
 * Parses an expression.
 * @param formula CSS nth-formula to parse.
 * @throws {Error} An `Error` if parsing fails.
 * @returns An array containing the integer step size and the integer offset of the nth rule.
 * @example nthCheck.parse("2n+3"); // returns [2, 3]
 */
func Parse(formula string) (a int, b int, err error) {
	formula = strings.ToLower(strings.TrimSpace(formula))

	if len(formula) == 0 {
		return 0, 0, errors.New("empty formula")
	}

	if formula == "even" {
		return 2, 0, nil
	}
	if formula == "odd" {
		return 2, 1, nil
	}

	// Parse [ ['-'|'+']? INTEGER? {N} [ S* ['-'|'+'] S* INTEGER ]?

	var index = 0

	readSign := func() int {
		switch formula[index] {
		case '-':
			{
				index++
				return -1
			}
		case '+':
			{
				index++
			}
		}

		return 1
	}

	readNumber := func() int {
		var start = index
		var value = 0

		for index < len(formula) &&
			formula[index] >= ZERO &&
			formula[index] <= NINE {
			value = value*10 + int(formula[index]-ZERO)
			index++
		}

		// Return `null` if we didn't read anything.
		if index == start {
			return -1
		}

		return value
	}

	skipWhitespace := func() {
		for index < len(formula) && isWhitespace(formula[index]) {
			index++
		}
	}

	a = 0
	var sign = readSign()
	var number = readNumber()

	if index < len(formula) && formula[index] == 'n' {
		index++
		if number != -1 {
			a = sign * number
		} else {
			a = sign
		}

		skipWhitespace()

		if index < len(formula) {
			sign = readSign()
			skipWhitespace()
			number = readNumber()
		} else {
			sign = 0
			number = 0
		}
	}

	// Throw if there is anything else
	if number == -1 || index < len(formula) {
		return 0, 0, fmt.Errorf("n-th rule couldn't be parsed ('%s')", formula)
	}

	return a, sign * number, nil
}
