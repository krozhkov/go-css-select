package helpers

import (
	"testing"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
 * Sorts the rules of a selector and turns it back into a string.
 *
 * Note that the order of the rules might not be legal, and the resulting
 * string might not be parseable again.
 * @param selector Selector to sort
 * @returns Sorted selector, which might not be a valid selector anymore.
 */
func parseSortStringify(t *testing.T, selector string) string {
	parsed, err := parser.Parse(selector)
	require.Nil(t, err)

	for _, token := range parsed {
		SortRules(token)
	}

	return parser.Stringify(parsed)
}

func TestSortRules(t *testing.T) {
	t.Run("should move tag selectors last", func(t *testing.T) {
		assert.Equal(t, ":empty[class]div", parseSortStringify(t, "div[class]:empty"))
	})

	t.Run("should move universal selectors last", func(t *testing.T) {
		assert.Equal(t, "[class]*", parseSortStringify(t, "*[class]"))
	})

	t.Run("should sort attribute selectors", func(t *testing.T) {
		assert.Equal(t,
			`.foo#bar[foo="bar" i][foo^="bar"][foo$="bar"][foo!="bar"][foo!="bar" s][foo="bar"]`,
			parseSortStringify(t,
				".foo#bar[foo=bar][foo^=bar][foo$=bar][foo!=bar][foo=bar i][foo!=bar s]",
			),
		)
	})

	t.Run("should sort pseudo selectors", func(t *testing.T) {
		assert.Equal(t,
			":contains(a):icontains(a):has(div):is(foo bar):not(:empty):empty:is([foo]):is(div)",
			parseSortStringify(t,
				":not(:empty):empty:contains(a):icontains(a):has(div):is(div):is(foo bar):is([foo])",
			),
		)
	})

	t.Run("should support traversals", func(t *testing.T) {
		assert.Equal(t,
			`div > :empty[foo]* + [bar="foo" i]:is(div)`,
			parseSortStringify(t, "div > *:empty[foo] + [bar=foo i]:is(div)"),
		)
	})
}
