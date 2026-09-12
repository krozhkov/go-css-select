package query

import (
	"testing"

	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/stretchr/testify/assert"
)

func TestAttributes(t *testing.T) {
	doc := parseDocument(
		`<div data-foo="In the end, it doesn't really matter."></div><div data-foo="Indeed-that's a delicate matter.">`,
	)
	children := doc.Children

	t.Run("ignore case", func(t *testing.T) {
		t.Run("should for =", func(t *testing.T) {
			matches, err := SelectAll(
				`[data-foo="indeed-that's a delicate matter." i]`,
				children,
				nil,
			)

			assert.Nil(t, err)

			assert.Len(t, matches, 1)
			assert.Equal(t, []*dom.Node{children[1]}, matches)

			matches, err = SelectAll(
				`[data-foo="inDeeD-THAT's a DELICATE matteR." i]`,
				children,
				nil,
			)
			assert.Equal(t, []*dom.Node{children[1]}, matches)
		})

		t.Run("should for ^=", func(t *testing.T) {
			matches, err := SelectAll("[data-foo^=IN i]", children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			assert.Equal(t, children, matches)
			matches, err = SelectAll("[data-foo^=in i]", children, nil)
			assert.Nil(t, err)
			assert.Equal(t, children, matches)
			matches, err = SelectAll("[data-foo^=iN i]", children, nil)
			assert.Nil(t, err)
			assert.Equal(t, children, matches)
		})

		t.Run("should for $=", func(t *testing.T) {
			matches, err := SelectAll(`[data-foo$="MATTER." i]`, children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			assert.Equal(t, children, matches)
			matches, err = SelectAll(`[data-foo$="matter." i]`, children, nil)
			assert.Nil(t, err)
			assert.Equal(t, children, matches)
			matches, err = SelectAll(`[data-foo$="MaTtEr." i]`, children, nil)
			assert.Nil(t, err)
			assert.Equal(t, children, matches)
		})

		t.Run("should for !=", func(t *testing.T) {
			matches, err := SelectAll(`[data-foo!="indeed-that's a delicate matter." i]`, children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, []*dom.Node{children[0]}, matches)
			matches, err = SelectAll(`[data-foo!="inDeeD-THAT's a DELICATE matteR." i]`, children, nil)
			assert.Nil(t, err)
			assert.Equal(t, []*dom.Node{children[0]}, matches)
		})

		t.Run("should for *=", func(t *testing.T) {
			matches, err := SelectAll("[data-foo*=IT i]", children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, []*dom.Node{children[0]}, matches)
			matches, err = SelectAll("[data-foo*=tH i]", children, nil)
			assert.Nil(t, err)
			assert.Equal(t, children, matches)
		})

		t.Run("should for |=", func(t *testing.T) {
			matches, err := SelectAll("[data-foo|=indeed i]", children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, []*dom.Node{children[1]}, matches)
			matches, err = SelectAll("[data-foo|=inDeeD i]", children, nil)
			assert.Nil(t, err)
			assert.Equal(t, []*dom.Node{children[1]}, matches)
		})

		t.Run("should for ~=", func(t *testing.T) {
			matches, err := SelectAll("[data-foo~=IT i]", children, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, []*dom.Node{children[0]}, matches)
			matches, err = SelectAll("[data-foo~=dElIcAtE i]", children, nil)
			assert.Nil(t, err)
			assert.Equal(t, []*dom.Node{children[1]}, matches)
		})
	})

	t.Run("no matches", func(t *testing.T) {
		t.Run("should for ~=", func(t *testing.T) {
			query, err := compileUnsafe("[foo~='baz bar']", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
		})

		t.Run("should for $=", func(t *testing.T) {
			query, err := compileUnsafe("[foo$='']", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
		})

		t.Run("should for *=", func(t *testing.T) {
			query, err := compileUnsafe("[foo*='']", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
		})
	})
}
