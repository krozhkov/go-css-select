package query

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCache(t *testing.T) {
	cache := helpers.NewCache[dom.Node, string]()

	node := dom.NewComment("test")

	cache.Set(node, "data")

	assert.Equal(t, 1, cache.Len())
	assert.True(t, cache.Has(node))
	assert.Equal(t, "data", cache.Get(node))

	node = nil
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, cache.Len())
}

func parseDocument(str string) *dom.Node {
	return dom.ParseDocument(str, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
}

type MatcherMock struct {
	mock.Mock
	match func(elem *dom.Node) bool
}

func (m *MatcherMock) Match(elem *dom.Node) bool {
	m.Called(elem)
	return m.match(elem)
}
func NewMatcherMock(match func(elem *dom.Node) bool) *MatcherMock {
	m := &MatcherMock{match: match}
	m.On("Match", mock.Anything).Return()
	return m
}

func TestCacheParentResults(t *testing.T) {
	t.Run("should rely on parent for matches", func(t *testing.T) {
		documentWithoutFoo := parseDocument(
			"<a><b><c><d><e>bar</e></d></c><f><g>bar</g></f></b></a>",
		)

		matcher := NewMatcherMock(func(elem *dom.Node) bool {
			text := domutils.GetText(elem)
			return strings.Contains(text, "foo")
		})
		fn := matcher.Match

		hasfoo := helpers.CacheParentResults(
			&types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return true
				},
			},
			nil,
			fn,
		)

		options := &types.Options{
			Pseudos: map[string]func(elem *dom.Node, value string) bool{
				"hasfoo": func(elem *dom.Node, _ string) bool {
					return hasfoo.Match(elem)
				},
			},
		}

		result, err := SelectAll(":hasfoo", documentWithoutFoo.Children, options)

		assert.Nil(t, err)

		assert.Len(t, result, 0)

		matcher.AssertNumberOfCalls(t, "Match", 1)
	})

	t.Run("should cache results for subtrees", func(t *testing.T) {
		documentWithFoo := parseDocument(
			"<a><b><c><d><e>foo</e></d></c><f><g>bar</g></f></b></a>",
		)

		matcher := NewMatcherMock(func(elem *dom.Node) bool {
			text := domutils.GetText(elem)
			return strings.Contains(text, "foo")
		})
		fn := matcher.Match

		hasfoo := helpers.CacheParentResults(
			&types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return true
				},
			},
			nil,
			fn,
		)

		options := &types.Options{
			Pseudos: map[string]func(elem *dom.Node, value string) bool{
				"hasfoo": func(elem *dom.Node, _ string) bool {
					return hasfoo.Match(elem)
				},
			},
		}

		result, err := SelectAll(":hasfoo", documentWithFoo.Children, options)

		assert.Nil(t, err)

		assert.Len(t, result, 5)

		matcher.AssertNumberOfCalls(t, "Match", 6)
	})

	t.Run("should use cached result for multiple matches", func(t *testing.T) {
		documentWithFoo := parseDocument(
			"<a><b><c><d><e>foo</e></d></c><f><g>bar</g></f></b></a>",
		)

		matcher := NewMatcherMock(func(elem *dom.Node) bool {
			text := domutils.GetText(elem)
			return strings.Contains(text, "foo")
		})
		fn := matcher.Match

		hasfoo := helpers.CacheParentResults(
			&types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return true
				},
			},
			nil,
			fn,
		)

		options := &types.Options{
			Pseudos: map[string]func(elem *dom.Node, value string) bool{
				"hasfoo": func(elem *dom.Node, _ string) bool {
					return hasfoo.Match(elem)
				},
			},
		}

		result, err := SelectAll(":hasfoo :hasfoo", documentWithFoo.Children, options)

		assert.Nil(t, err)

		assert.Len(t, result, 4)

		matcher.AssertNumberOfCalls(t, "Match", 6)
	})
}
