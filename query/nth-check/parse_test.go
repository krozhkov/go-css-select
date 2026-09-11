package nthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var valid = map[string][2]int{
	"1":       {0, 1},
	"2":       {0, 2},
	"3":       {0, 3},
	"5":       {0, 5},
	" 1 ":     {0, 1},
	" 5 ":     {0, 5},
	"+2n + 1": {2, 1},
	"-1":      {0, -1},
	"-1n + 3": {-1, 3},
	"-1n+3":   {-1, 3},
	"-n+2":    {-1, 2},
	"-n+3":    {-1, 3},
	"0n+3":    {0, 3},
	"1n":      {1, 0},
	"1n+0":    {1, 0},
	"2n":      {2, 0},
	"2n + 1":  {2, 1},
	"2n+1":    {2, 1},
	"3n":      {3, 0},
	"3n+0":    {3, 0},
	"3n+1":    {3, 1},
	"3n+2":    {3, 2},
	"3n+3":    {3, 3},
	"3n-1":    {3, -1},
	"3n-2":    {3, -2},
	"3n-3":    {3, -3},
	"even":    {2, 0},
	"n":       {1, 0},
	"n+2":     {1, 2},
	"odd":     {2, 1},

	// Surprisingly, neither sizzle, qwery or nwmatcher cover these cases
	"-4n+13":   {-4, 13},
	"-2n + 12": {-2, 12},
	"-n":       {-1, 0},
}

var invalid = []string{
	"-",
	"- 1n",
	"-1 n",
	"2+0",
	"2n+-0",
	"an+b",
	"asdf",
	"b",
	"expr",
	"odd|even|x",
}

func TestParse(t *testing.T) {
	t.Run("parse invalid", func(t *testing.T) {
		for _, formula := range invalid {
			_, _, err := Parse(formula)
			require.NotNil(t, err)
		}
	})

	t.Run("parse valid", func(t *testing.T) {
		for formula, result := range valid {
			a, b, err := Parse(formula)
			require.Nil(t, err)

			assert.Equal(t, a, result[0])
			assert.Equal(t, b, result[1])
		}
	})
}
