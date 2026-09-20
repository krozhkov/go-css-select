package pseudoselectors

import (
	"slices"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

type Subselect = func(
	next *types.CompiledQuery,
	subselect [][]*parser.Selector,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
) (*types.CompiledQuery, error)

/**
 * Check if the selector has any properties that rely on the current element.
 * If not, we can cache the result of the selector.
 *
 * We can't cache selectors that start with a traversal (e.g. `>`, `+`, `~`),
 * or include a `:scope`.
 * @param selector - The selector to check.
 * @returns Whether the selector has any properties that rely on the current element.
 */
func hasDependsOnCurrentElement(selector [][]*parser.Selector) bool {
	return slices.IndexFunc(selector, func(sel []*parser.Selector) bool {
		return len(sel) > 0 && (helpers.IsTraversal(sel[0]) || slices.IndexFunc(sel, helpers.IncludesScopePseudo) >= 0)
	}) >= 0
}

func is(
	next *types.CompiledQuery,
	token [][]*parser.Selector,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
) (*types.CompiledQuery, error) {
	compiledToken, err := compileToken(token, helpers.CopyOptions(options), context)
	if err != nil {
		return nil, err
	}

	if compiledToken.Type == types.MatchTypeAlwaysTrue {
		return next, nil
	}
	if compiledToken.Type == types.MatchTypeAlwaysFalse {
		return compiledToken, nil
	}

	return &types.CompiledQuery{
		Match: func(element *dom.Node, scope *dom.Node) bool {
			return compiledToken.Match(element, scope) && next.Match(element, scope)
		},
	}, nil
}

/*
 * :not, :has, :is, :matches and :where have to compile selectors
 * doing this in src/pseudos.ts would lead to circular dependencies,
 * so we add them here
 */
/** Pseudo selectors that compile nested selectors. */
var subselects = map[string]Subselect{
	"is": is,
	/**
	 * `:matches` and `:where` are aliases for `:is`.
	 */
	"matches": is,
	"where":   is,
	"not": func(
		next *types.CompiledQuery,
		token [][]*parser.Selector,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		compiledToken, err := compileToken(token, helpers.CopyOptions(options), context)
		if err != nil {
			return nil, err
		}

		if compiledToken.Type == types.MatchTypeAlwaysFalse {
			return next, nil
		}
		if compiledToken.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return !compiledToken.Match(elem, scope) && next.Match(elem, scope)
			},
		}, nil
	},
	"has": func(
		next *types.CompiledQuery,
		subselect [][]*parser.Selector,
		options *types.Options,
		_ []*dom.Node,
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		copiedOptions := helpers.CopyOptions(options)
		copiedOptions.RelativeSelector = types.OptYes

		compiled, err := compileToken(subselect, copiedOptions, nil)
		if err != nil {
			return nil, err
		}

		if compiled.Type == types.MatchTypeAlwaysFalse {
			return compiled, nil
		}

		skipCache := hasDependsOnCurrentElement(subselect)

		// If `compiled` is `trueFunc`, we can use this.
		if compiled.Type == types.MatchTypeAlwaysTrue {
			hasOne := func(element *dom.Node, scope *dom.Node) bool {
				return helpers.FindOne(compiled.Match, domutils.GetChildren(element), scope, options) != nil
			}

			if skipCache {
				return &types.CompiledQuery{
					Match: func(elem *dom.Node, scope *dom.Node) bool {
						return next.Match(elem, scope) && hasOne(elem, scope)
					},
				}, nil
			} else {
				return helpers.CacheParentResults(next, options, hasOne), nil
			}
		}

		hasMatch := func(element *dom.Node, scope *dom.Node) bool {
			children := domutils.GetChildren(element)

			if compiled.ShouldTestNextSiblings {
				nextSiblings := helpers.GetNextSiblings(element)
				children = slices.Grow(children, len(nextSiblings))
				children = append(children, nextSiblings...)
			}

			matchForScope := func(node *dom.Node, _ *dom.Node) bool {
				return compiled.Match(node, element) // pass "elem" as a new scope
			}

			return helpers.FindOne(matchForScope, children, scope, options) != nil
		}

		if skipCache {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return next.Match(element, scope) && hasMatch(element, scope)
				},
			}, nil
		}

		return helpers.CacheParentResults(next, options, hasMatch), nil
	},
}
