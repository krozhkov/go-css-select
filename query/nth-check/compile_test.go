package nthcheck

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
 * Iterate through all possible values. This is adapted from qwery,
 * and uses a more intuitive way to process all elements.
 * @param rule A tuple [a, b] representing an nth-rule, as returned by `parse`.
 * @param rule."0" The `a` value from the tuple.
 * @param rule."1" The `b` value from the tuple.
 */
func slowNth(valueArray []int, a, b int) []int {
	if a == 0 && b > 0 {
		return []int{b - 1}
	}

	compare := func(a int, index int, length int) bool {
		if a > 0 {
			return index <= length
		} else {
			return index >= 1
		}
	}

	length := len(valueArray)

	return slowFilter(valueArray, func(value int) bool {
		for index := b; compare(a, index, length); index += a {
			if index-1 >= 0 && index-1 < len(valueArray) && value == valueArray[index-1] {
				return true
			}
		}
		return false
	})
}

func slowFilter(values []int, predicate func(value int) bool) []int {
	var filtered []int
	for _, value := range values {
		if predicate(value) {
			filtered = append(filtered, value)
		}
	}

	return filtered
}

func TestCompile(t *testing.T) {
	valueArray := make([]int, 2e3)
	for i := range valueArray {
		valueArray[i] = i
	}

	t.Run("compile & run all valid", func(t *testing.T) {
		for rule, parsed := range valid {
			t.Run(fmt.Sprintf("compile & run all valid for %s", rule), func(t *testing.T) {
				check := Compile(parsed[0], parsed[1])
				filtered := slowFilter(valueArray, check.Fn)
				iterated := slowNth(valueArray, parsed[0], parsed[1])

				assert.Equal(t, filtered, iterated)
			})
		}
	})

	t.Run("parse, compile & run all valid", func(t *testing.T) {
		for rule, parsed := range valid {
			t.Run(fmt.Sprintf("parse, compile & run all valid for %s", rule), func(t *testing.T) {
				a, b, err := Parse(rule)
				require.Nil(t, err)
				check := Compile(a, b)
				filtered := slowFilter(valueArray, check.Fn)
				iterated := slowNth(valueArray, parsed[0], parsed[1])

				assert.Equal(t, filtered, iterated)
			})
		}
	})
}

func TestGenerate(t *testing.T) {
	t.Run("should only return valid values", func(t *testing.T) {
		for _, parsed := range valid {
			gen := Generate(parsed[0], parsed[1])
			check := Compile(parsed[0], parsed[1])
			value := gen()

			for index := 0; index < 1e3; index++ {
				// Should pass the check iff `i` is the next value.
				assert.Equal(t, value == index, check.Fn(index))

				if value == index {
					value = gen()
				}
			}
		}
	})

	t.Run("should produce an increasing sequence", func(t *testing.T) {
		gen := Generate(2, 2)

		assert.Equal(t, gen(), 1)
		assert.Equal(t, gen(), 3)
		assert.Equal(t, gen(), 5)
		assert.Equal(t, gen(), 7)
		assert.Equal(t, gen(), 9)
	})

	t.Run("should produce an increasing sequence for a negative `n`", func(t *testing.T) {
		gen := Generate(-1, 2)

		assert.Equal(t, gen(), 0)
		assert.Equal(t, gen(), 1)
		assert.Equal(t, gen(), -1)
	})

	t.Run("should not produce any values for `-n`", func(t *testing.T) {
		gen := Generate(-1, 0)

		assert.Equal(t, gen(), -1)
	})

	t.Run("should parse selectors with `sequence`", func(t *testing.T) {
		a, b, err := Parse("-2n+5")
		require.Nil(t, err)

		gen := Generate(a, b)

		assert.Equal(t, gen(), 0)
		assert.Equal(t, gen(), 2)
		assert.Equal(t, gen(), 4)
		assert.Equal(t, gen(), -1)
	})
}
