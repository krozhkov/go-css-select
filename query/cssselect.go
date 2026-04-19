package query

import (
	"slices"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

func convertOptionFormats(
	options *types.Options,
) *types.Options {
	if options == nil {
		options = &types.Options{}
	}

	return options
}

/**
 * Compiles a selector to an executable function.
 *
 * The returned function checks if each passed node is an element. Use
 * `_compileUnsafe` to skip this check.
 *
 * @param selector Selector to compile.
 * @param options Compilation options.
 * @param context Optional context for the selector.
 */
func Compile(
	selector string,
	options *types.Options,
	context []*dom.Node,
) (func(*dom.Node) bool, error) {
	opts := convertOptionFormats(options)
	next, err := compileUnsafe(selector, opts, context)
	if err != nil {
		return nil, err
	}

	if next.Type == types.MatchTypeAlwaysFalse {
		return next.Match, nil
	}

	return func(elem *dom.Node) bool {
		return dom.IsTag(elem) && next.Match(elem)
	}, nil
}

/**
 * Like `compile`, but does not add a check if elements are tags.
 */
func compileUnsafe(
	selector string,
	options *types.Options,
	context []*dom.Node,
) (*types.CompiledQuery, error) {
	token, err := parser.Parse(selector)
	if err != nil {
		return nil, err
	}

	return compileToken(token, options, context)
}

func prepareContext(
	elems []*dom.Node,
	shouldTestNextSiblings bool,
) []*dom.Node {
	/*
	 * Add siblings if the query requires them.
	 * See https://github.com/fb55/css-select/pull/43#issuecomment-225414692
	 */
	if shouldTestNextSiblings {
		elems = appendNextSiblings(elems)
	}

	return domutils.RemoveSubsets(elems)
}

func appendNextSiblings(
	elems []*dom.Node,
) []*dom.Node {
	elemsLength := len(elems)
	for i := 0; i < elemsLength; i++ {
		nextSiblings := helpers.GetNextSiblings(elems[i])
		elems = slices.Concat(elems, nextSiblings)
	}
	return elems
}

/**
 * @template Node The generic Node type for the DOM adapter being used.
 * @template ElementNode The Node type for elements for the DOM adapter being used.
 * @param elems Elements to query. If it is an element, its children will be queried.
 * @param query can be either a CSS selector string or a compiled query function.
 * @param [options] options for querying the document.
 * @see compile for supported selector queries.
 * @returns All matching elements.
 *
 */
func SelectAll(
	selector string,
	elements []*dom.Node,
	options *types.Options,
) ([]*dom.Node, error) {
	opts := convertOptionFormats(options)
	query, err := compileUnsafe(selector, opts, elements)
	if err != nil {
		return nil, err
	}

	filteredElements := prepareContext(
		elements,
		query.ShouldTestNextSiblings,
	)

	if query.Type == types.MatchTypeAlwaysFalse || len(filteredElements) == 0 {
		return nil, nil
	}

	return helpers.FindAll(query.Match, filteredElements, opts), nil
}

/**
 * @template Node The generic Node type for the DOM adapter being used.
 * @template ElementNode The Node type for elements for the DOM adapter being used.
 * @param elems Elements to query. If it is an element, its children will be queried.
 * @param query can be either a CSS selector string or a compiled query function.
 * @param [options] options for querying the document.
 * @see compile for supported selector queries.
 * @returns the first match, or null if there was no match.
 */
func SelectOne(
	selector string,
	elements []*dom.Node,
	options *types.Options,
) (*dom.Node, error) {
	opts := convertOptionFormats(options)
	query, err := compileUnsafe(selector, opts, elements)
	if err != nil {
		return nil, err
	}

	filteredElements := prepareContext(
		elements,
		query.ShouldTestNextSiblings,
	)

	if query.Type == types.MatchTypeAlwaysFalse || len(filteredElements) == 0 {
		return nil, nil
	}

	return helpers.FindOne(query.Match, filteredElements, opts), nil
}
