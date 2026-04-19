package pseudoselectors

import (
	"slices"
	"strings"

	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

type Filter = func(
	next *types.CompiledQuery,
	text string,
	options *types.Options,
	context []*dom.Node,
) *types.CompiledQuery

var filters = map[string]Filter{
	"contains": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) *types.CompiledQuery {
		return helpers.CacheParentResults(next, options, func(elem *dom.Node) bool {
			return strings.Contains(domutils.GetText(elem), text)
		})
	},
	"icontains": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) *types.CompiledQuery {
		itext := strings.ToLower(text)

		return helpers.CacheParentResults(next, options, func(elem *dom.Node) bool {
			return strings.Contains(strings.ToLower(domutils.GetText(elem)), itext)
		})
	},
	// Location specific methods
	// "nth-child" - not supported
	// "nth-last-child" - not supported
	// "nth-of-type" - not supported
	// "nth-last-of-type" - not supported

	// TODO determine the actual root element
	"root": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) *types.CompiledQuery {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return helpers.GetElementParent(elem) == nil && next.Match(elem)
			},
		}
	},

	"scope": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) *types.CompiledQuery {
		if len(context) == 0 {
			// Equivalent to :root
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return helpers.GetElementParent(elem) == nil && next.Match(elem)
				},
			}
		}

		if len(context) == 1 {
			// NOTE: can't be unpacked, as :has uses this for side-effects
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return context[0] == elem && next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return slices.Index(context, elem) >= 0 && next.Match(elem)
			},
		}
	},

	// "hover" - not supported
	// "visited" - not supported
	// "active" - not supported
}
