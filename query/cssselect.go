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
 * @param selector Selector to compile.
 * @param options Compilation options.
 * @param context Optional context for the selector.
 */
func Compile[T *dom.Node | []*dom.Node](
	selector string,
	options *types.Options,
	context T,
) (func(*dom.Node) bool, error) {
	convertedOptions := convertOptionFormats(options)
	next, err := compileUnsafe(selector, convertedOptions, context)
	if err != nil {
		return nil, err
	}

	if next.Type == types.MatchTypeAlwaysFalse {
		return func(elem *dom.Node) bool {
			return next.Match(elem, nil)
		}, nil
	}

	return func(element *dom.Node) bool {
		return dom.IsTag(element) && next.Match(element, nil)
	}, nil
}

/**
 * Like `compile`, but does not add a check if elements are tags.
 * @param selector Selector used to match elements.
 * @param options Options that control this operation.
 * @param context Context nodes used to scope selector matching.
 */
func compileUnsafe[T *dom.Node | []*dom.Node](
	selector string,
	options *types.Options,
	context T,
) (*types.CompiledQuery, error) {
	token, err := parser.Parse(selector)
	if err != nil {
		return nil, err
	}

	return compileToken(token, options, context)
}

/**
 * Normalize a query context and optionally include next siblings.
 * @param element Elements to test against sibling-dependent selectors.
 * @param adapter Adapter implementation used for DOM operations.
 * @param shouldTestNextSiblings Whether sibling combinators should include following siblings.
 */
func prepareContext[T *dom.Node | []*dom.Node](
	element T,
	shouldTestNextSiblings bool,
) []*dom.Node {
	switch v := any(element).(type) {
	case *dom.Node:
		{
			/*
			 * Add siblings if the query requires them.
			 * See https://github.com/fb55/css-select/pull/43#issuecomment-225414692
			 */
			if shouldTestNextSiblings {
				elements := appendNextSiblings(v)

				return domutils.RemoveSubsets(elements)
			}

			return domutils.GetChildren(v)
		}
	case []*dom.Node:
		{
			elements := v
			/*
			 * Add siblings if the query requires them.
			 * See https://github.com/fb55/css-select/pull/43#issuecomment-225414692
			 */
			if shouldTestNextSiblings {
				elements = appendNextSiblings(elements)
			}

			return domutils.RemoveSubsets(elements)
		}
	default:
		return nil
	}
}

func appendNextSiblings[T *dom.Node | []*dom.Node](
	element T,
) []*dom.Node {
	var elements []*dom.Node
	switch v := any(element).(type) {
	case *dom.Node:
		elements = append(elements, v)
	case []*dom.Node:
		elements = slices.Clone(v)
	}

	elementsLength := len(elements)
	for i := 0; i < elementsLength; i++ {
		nextSiblings := helpers.GetNextSiblings(elements[i])
		elements = slices.Grow(elements, len(nextSiblings))
		elements = append(elements, nextSiblings...)
	}
	return elements
}

/**
 * @template Node The generic Node type for the DOM adapter being used.
 * @template ElementNode The Node type for elements for the DOM adapter being used.
 * @param elements Elements to query. If it is an element, its children will be queried.
 * @param query can be either a CSS selector string or a compiled query function.
 * @param [options] options for querying the document.
 * @see compile for supported selector queries.
 * @returns All matching elements.
 */
func SelectAll[T *dom.Node | []*dom.Node](
	selector string,
	elements T,
	options *types.Options,
) ([]*dom.Node, error) {
	convertedOptions := convertOptionFormats(options)
	query, err := compileUnsafe(selector, convertedOptions, elements)
	if err != nil {
		return nil, err
	}

	filteredElements := prepareContext(
		elements,
		query.ShouldTestNextSiblings,
	)

	if query.Type == types.MatchTypeAlwaysFalse || len(filteredElements) == 0 {
		return []*dom.Node{}, nil
	}

	return helpers.FindAll(query.Match, filteredElements, nil, convertedOptions), nil
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
func SelectOne[T *dom.Node | []*dom.Node](
	selector string,
	elements T,
	options *types.Options,
) (*dom.Node, error) {
	convertedOptions := convertOptionFormats(options)
	query, err := compileUnsafe(selector, convertedOptions, elements)
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

	return helpers.FindOne(query.Match, filteredElements, nil, convertedOptions), nil
}

/**
 * Tests whether or not an element is matched by query.
 * @template Node The generic Node type for the DOM adapter being used.
 * @template ElementNode The Node type for elements for the DOM adapter being used.
 * @param element The element to test if it matches the query.
 * @param query can be either a CSS selector string or a compiled query function.
 * @param [options] options for querying the document.
 * @see compile for supported selector queries.
 * @returns Whether the element matches the query.
 */
func Is(
	element *dom.Node,
	query string,
	options *types.Options,
) (bool, error) {
	compiled, err := Compile[*dom.Node](query, options, nil)
	if err != nil {
		return false, err
	}

	return compiled(element), nil
}
