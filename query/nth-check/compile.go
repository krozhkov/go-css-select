package nthcheck

import (
	"github.com/krozhkov/go-css-select/query/types"
)

type CompiledNthCheck struct {
	Fn   func(idx int) bool
	Type types.MatchType
}

/**
 * Returns a function that checks if an elements index matches the given rule
 * highly optimized to return the fastest solution.
 * @param parsed A tuple [a, b], as returned by `parse`.
 * @returns A highly optimized function that returns whether an index matches the nth-check.
 * @example
 *
 * ```js
 * const check = nthCheck.compile([2, 3]);
 *
 * check(0); // `false`
 * check(1); // `false`
 * check(2); // `true`
 * check(3); // `false`
 * check(4); // `true`
 * check(5); // `false`
 * check(6); // `true`
 * ```
 */
func Compile(a int, b int) *CompiledNthCheck {
	// Subtract 1 from `b`, to convert from one- to zero-indexed.
	b = b - 1

	/*
	 * When `b <= 0`, `a * n` won't be lead to any matches for `a < 0`.
	 * Besides, the specification states that no elements are
	 * matched when `a` and `b` are 0.
	 *
	 * `b < 0` here as we subtracted 1 from `b` above.
	 */
	if b < 0 && a <= 0 {
		return &CompiledNthCheck{
			Fn: func(index int) bool {
				return false
			},
			Type: types.MatchTypeAlwaysFalse,
		}
	}

	// When `a` is in the range -1..1, it matches any element (so only `b` is checked).
	if a == -1 {
		return &CompiledNthCheck{
			Fn: func(index int) bool {
				return index <= b
			},
		}
	}
	if a == 0 {
		return &CompiledNthCheck{
			Fn: func(index int) bool {
				return index == b
			},
		}
	}
	// When `b <= 0` and `a === 1`, they match any element.
	if a == 1 {
		if b < 0 {
			return &CompiledNthCheck{
				Fn: func(index int) bool {
					return true
				},
				Type: types.MatchTypeAlwaysTrue,
			}
		}

		return &CompiledNthCheck{
			Fn: func(index int) bool {
				return index >= b
			},
		}
	}

	/*
	 * Otherwise, modulo can be used to check if there is a match.
	 *
	 * Modulo doesn't care about the sign, so let's use `a`s absolute value.
	 */
	absA := a
	if a < 0 {
		absA = -a
	}
	// Get `b mod a`, + a if this is negative.
	bModulo := ((b % absA) + absA) % absA

	if a > 1 {
		return &CompiledNthCheck{
			Fn: func(index int) bool {
				return index >= b && index%absA == bModulo
			},
		}
	}

	return &CompiledNthCheck{
		Fn: func(index int) bool {
			return index <= b && index%absA == bModulo
		},
	}
}

/**
 * Returns a function that produces a monotonously increasing sequence of indices.
 *
 * If the sequence has an end, the returned function will return `null` after
 * the last index in the sequence.
 * @param parsed A tuple [a, b], as returned by `parse`.
 * @returns A function that produces a sequence of indices.
 * @example <caption>Always increasing (2n+3)</caption>
 *
 * ```js
 * const gen = nthCheck.generate([2, 3])
 *
 * gen() // `1`
 * gen() // `3`
 * gen() // `5`
 * gen() // `8`
 * gen() // `11`
 * ```
 * @example <caption>With end value (-2n+10)</caption>
 *
 * ```js
 *
 * const gen = nthCheck.generate([-2, 5]);
 *
 * gen() // 0
 * gen() // 2
 * gen() // 4
 * gen() // null
 * ```
 */
func Generate(a int, b int) func() int {
	// Subtract 1 from `b`, to convert from one- to zero-indexed.
	b = b - 1

	n := 0

	// Make sure to always return an increasing sequence
	if a < 0 {
		aPos := -a
		// Get `b mod a`
		minValue := ((b % aPos) + aPos) % aPos
		return func() int {
			value := minValue + aPos*n
			n++

			if value > b {
				return -1
			}

			return value
		}
	}

	if a == 0 {
		if b < 0 {
			// There are no result — always return `null`
			return func() int {
				return -1
			}
		}
		// Return `b` exactly once
		return func() int {
			var result int
			if n == 0 {
				result = b
			} else {
				result = -1
			}
			n++
			return result
		}
	}

	if b < 0 {
		b += a * ((-b + a - 1) / a)
	}

	return func() int {
		result := a*n + b
		n++
		return result
	}
}
