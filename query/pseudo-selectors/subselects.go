package pseudoselectors

import (
	"slices"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

/** Used as a placeholder for :has. Will be replaced with the actual element. */
var PLACEHOLDER_ELEMENT = &dom.Node{}

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
 *
 * @param selector - The selector to check.
 * @returns Whether the selector has any properties that rely on the current element.
 */
func hasDependsOnCurrentElement(selector [][]*parser.Selector) bool {
	return slices.IndexFunc(selector, func(sel []*parser.Selector) bool {
		return len(sel) > 0 && (helpers.IsTraversal(sel[0]) || slices.IndexFunc(sel, helpers.IncludesScopePseudo) >= 0)
	}) >= 0
}

func ptr[T any](v T) *T {
	return &v
}

func copyOptions(
	options *types.Options,
) *types.Options {
	if options == nil {
		return nil
	}

	var xmlMode = types.OptNo
	if options.XmlMode == types.OptYes {
		xmlMode = types.OptYes
	}
	var lowerCaseAttributeNames = types.OptNo
	if options.LowerCaseAttributeNames == types.OptYes {
		lowerCaseAttributeNames = types.OptYes
	}
	var lowerCaseTags = types.OptNo
	if options.LowerCaseTags == types.OptYes {
		lowerCaseTags = types.OptYes
	}
	var quirksMode = types.OptNo
	if options.QuirksMode == types.OptYes {
		quirksMode = types.OptYes
	}
	var cacheResults = types.OptNo
	if options.CacheResults == types.OptYes {
		cacheResults = types.OptYes
	}
	// Not copied: context, rootFunc
	return &types.Options{
		XmlMode:                 xmlMode,
		LowerCaseAttributeNames: lowerCaseAttributeNames,
		LowerCaseTags:           lowerCaseTags,
		QuirksMode:              quirksMode,
		CacheResults:            cacheResults,
		Pseudos:                 options.Pseudos,
	}
}

func is(
	next *types.CompiledQuery,
	token [][]*parser.Selector,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
) (*types.CompiledQuery, error) {
	query, err := compileToken(token, copyOptions(options), context)
	if err != nil {
		return nil, err
	}

	if query.Type == types.MatchTypeAlwaysTrue {
		return next, nil
	}
	if query.Type == types.MatchTypeAlwaysFalse {
		return query, nil
	}

	return &types.CompiledQuery{
		Match: func(elem *dom.Node) bool {
			return query.Match(elem) && next.Match(elem)
		},
	}, nil
}

/*
 * :not, :has, :is, :matches and :where have to compile selectors
 * doing this in src/pseudos.ts would lead to circular dependencies,
 * so we add them here
 */
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
		query, err := compileToken(token, copyOptions(options), context)
		if err != nil {
			return nil, err
		}

		if query.Type == types.MatchTypeAlwaysFalse {
			return next, nil
		}
		if query.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return !query.Match(elem) && next.Match(elem)
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
		opts := copyOptions(options)
		opts.RelativeSelector = types.OptYes

		var context []*dom.Node
		if slices.IndexFunc(subselect, func(s []*parser.Selector) bool {
			return slices.IndexFunc(s, helpers.IsTraversal) >= 0
		}) >= 0 {
			context = []*dom.Node{PLACEHOLDER_ELEMENT}
		}
		skipCache := hasDependsOnCurrentElement(subselect)

		compiled, err := compileToken(subselect, opts, context)
		if err != nil {
			return nil, err
		}

		if compiled.Type == types.MatchTypeAlwaysFalse {
			return compiled, nil
		}

		// If `compiled` is `trueFunc`, we can skip this.
		if len(context) > 0 && compiled.Type != types.MatchTypeAlwaysTrue {
			if skipCache {
				return &types.CompiledQuery{
					Match: func(elem *dom.Node) bool {
						if !next.Match(elem) {
							return false
						}

						context[0] = elem
						childs := domutils.GetChildren(elem)

						if compiled.ShouldTestNextSiblings {
							childs = slices.Concat(childs, helpers.GetNextSiblings(elem))
						}

						return helpers.FindOne(compiled.Match, childs, options) != nil
					},
				}, nil
			} else {
				return helpers.CacheParentResults(next, options, func(elem *dom.Node) bool {
					context[0] = elem

					return helpers.FindOne(compiled.Match, domutils.GetChildren(elem), options) != nil
				}), nil
			}
		}

		hasOne := func(elem *dom.Node) bool {
			return helpers.FindOne(compiled.Match, domutils.GetChildren(elem), options) != nil
		}

		if skipCache {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return next.Match(elem) && hasOne(elem)
				},
			}, nil
		} else {
			return helpers.CacheParentResults(next, options, hasOne), nil
		}
	},
}
