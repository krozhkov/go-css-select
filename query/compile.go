package query

import (
	"errors"
	"slices"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/internal"
	pseudoselectors "github.com/krozhkov/go-css-select/query/pseudo-selectors"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
)

var DESCENDANT_TOKEN = &parser.Selector{Type: parser.SelectorTypeDescendant}
var FLEXIBLE_DESCENDANT_TOKEN = &parser.Selector{
	Type: "_flexibleDescendant",
}
var SCOPE_TOKEN = &parser.Selector{
	Type: parser.SelectorTypePseudo,
	Name: "scope",
}

/*
 * CSS 4 Spec (Draft): 3.4.1. Absolutizing a Relative Selector
 * http://www.w3.org/TR/selectors4/#absolutizing
 */
func absolutize(
	token [][]*parser.Selector,
	context []*dom.Node,
) {
	// TODO Use better check if the context is a document
	hasContext := context != nil && internal.Every(context, func(e *dom.Node) bool {
		return e == pseudoselectors.PLACEHOLDER_ELEMENT || (dom.IsTag(e) && helpers.GetElementParent(e) != nil)
	})

	for index, t := range token {
		if len(t) > 0 && helpers.IsTraversal(t[0]) && t[0].Type != parser.SelectorTypeDescendant {
			// Don't continue in else branch
		} else if hasContext && slices.IndexFunc(t, helpers.IncludesScopePseudo) == -1 {
			t = slices.Insert(t, 0, DESCENDANT_TOKEN)
			token[index] = t
		} else {
			continue
		}

		token[index] = slices.Insert(t, 0, SCOPE_TOKEN)
	}
}

func or(a *types.CompiledQuery, b *types.CompiledQuery) *types.CompiledQuery {
	return &types.CompiledQuery{
		Match: func(node *dom.Node) bool {
			return a.Match(node) || b.Match(node)
		},
	}
}

func compileToken(
	token [][]*parser.Selector,
	options *types.Options,
	context []*dom.Node,
) (*types.CompiledQuery, error) {
	for _, t := range token {
		helpers.SortRules(t)
	}

	finalContext := context
	if options != nil && options.Context != nil {
		finalContext = options.Context
	}
	rootFunc := &types.CompiledQuery{
		Match: func(element *dom.Node) bool {
			return true
		},
		Type: types.MatchTypeAlwaysTrue,
	}
	if options != nil && options.RootFunc != nil {
		rootFunc.Match = options.RootFunc
		rootFunc.Type = types.MatchTypeUnknown
	}

	// Check if the selector is relative
	if options == nil || options.RelativeSelector != types.OptNo {
		absolutize(token, finalContext)
	} else if slices.IndexFunc(token, func(t []*parser.Selector) bool { return len(t) > 0 && helpers.IsTraversal(t[0]) }) >= 0 {
		return nil, errors.New("relative selectors are not allowed when the `relativeSelector` option is disabled")
	}

	shouldTestNextSiblings := false
	query := &types.CompiledQuery{
		Match: func(elem *dom.Node) bool {
			return false
		},
		Type: types.MatchTypeAlwaysFalse,
	}

combineLoop:
	for _, rules := range token {
		if len(rules) >= 2 {
			first := rules[0]
			second := rules[1]

			if first.Type != parser.SelectorTypePseudo || first.Name != "scope" {
				// Ignore
			} else if second.Type == parser.SelectorTypeDescendant {
				rules[1] = FLEXIBLE_DESCENDANT_TOKEN
			} else if second.Type == parser.SelectorTypeAdjacent || second.Type == parser.SelectorTypeSibling {
				shouldTestNextSiblings = true
			}
		}

		next := rootFunc
		hasExpensiveSubselector := false
		var err error

		for _, rule := range rules {
			next, err = CompileGeneralSelector(
				next,
				rule,
				options,
				finalContext,
				compileToken,
				hasExpensiveSubselector,
			)

			if err != nil {
				return nil, err
			}

			quality := helpers.GetQuality(rule)

			if quality == 0 {
				hasExpensiveSubselector = true
			}

			// If the sub-selector won't match any elements, skip it.
			if next.Type == types.MatchTypeAlwaysFalse {
				continue combineLoop
			}
		}

		// If we have a function that always returns true, we can stop here.
		if next.Type == types.MatchTypeAlwaysTrue {
			return next, nil
		}

		query = or(query, next)
	}

	query.ShouldTestNextSiblings = shouldTestNextSiblings

	return query, nil
}
