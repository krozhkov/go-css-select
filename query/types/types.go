package types

import (
	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-htmlparser2/dom"
)

type OptBool int

const (
	OptUnspecified OptBool = iota
	OptYes
	OptNo
)

type Options struct {
	/**
	 * When enabled, tag names will be case-sensitive.
	 *
	 * @default false
	 */
	XmlMode OptBool
	/**
	 * Lower-case attribute names.
	 *
	 * @default !xmlMode
	 */
	LowerCaseAttributeNames OptBool
	/**
	 * Lower-case tag names.
	 *
	 * @default !xmlMode
	 */
	LowerCaseTags OptBool
	/**
	 * Is the document in quirks mode?
	 *
	 * This will lead to .className and #id being case-insensitive.
	 *
	 * @default false
	 */
	QuirksMode OptBool
	/**
	 * Pseudo-classes that override the default ones.
	 *
	 * Maps from names to either strings of functions.
	 * - A string value is a selector that the element must match to be selected.
	 * - A function is called with the element as its first argument, and optional
	 *  parameters second. If it returns true, the element is selected.
	 */
	Pseudos map[string]func(elem *dom.Node, value string) bool
	/**
	 * The last function in the stack, will be called with the last element
	 * that's looked at.
	 */
	RootFunc func(element *dom.Node) bool
	/**
	 * The context of the current query. Used to limit the scope of searches.
	 * Can be matched directly using the `:scope` pseudo-class.
	 */
	Context []*dom.Node
	/**
	 * Indicates whether to consider the selector as a relative selector.
	 *
	 * Relative selectors that don't include a `:scope` pseudo-class behave
	 * as if they have a `:scope ` prefix (a `:scope` pseudo-class, followed by
	 * a descendant selector).
	 *
	 * If relative selectors are disabled, selectors starting with a traversal
	 * will lead to an error.
	 *
	 * @default true
	 * @see {@link https://www.w3.org/TR/selectors-4/#relative}
	 */
	RelativeSelector OptBool
	/**
	 * Allow css-select to cache results for some selectors, sometimes greatly
	 * improving querying performance. Disable this if your document can
	 * change in between queries with the same compiled selector.
	 *
	 * @default true
	 */
	CacheResults OptBool
}

type MatchType int

const (
	MatchTypeUnknown MatchType = iota
	MatchTypeAlwaysTrue
	MatchTypeAlwaysFalse
)

type CompiledQuery struct {
	Match                  func(node *dom.Node) bool
	ShouldTestNextSiblings bool
	Type                   MatchType
}

type CompileToken func(token [][]*parser.Selector, options *Options, context []*dom.Node) (*CompiledQuery, error)
