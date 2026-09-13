package pseudoselectors

import (
	"slices"
	"strings"

	"github.com/krozhkov/go-css-select/query/helpers"
	nthcheck "github.com/krozhkov/go-css-select/query/nth-check"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

type Filter = func(
	next *types.CompiledQuery,
	text string,
	options *types.Options,
	context []*dom.Node,
) (*types.CompiledQuery, error)

var filters = map[string]Filter{
	"contains": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		return helpers.CacheParentResults(next, options, func(elem *dom.Node, scope *dom.Node) bool {
			return strings.Contains(domutils.GetText(elem), text)
		}), nil
	},
	"icontains": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		itext := strings.ToLower(text)

		return helpers.CacheParentResults(next, options, func(elem *dom.Node, scope *dom.Node) bool {
			return strings.Contains(strings.ToLower(domutils.GetText(elem)), itext)
		}), nil
	},
	// Location specific methods
	"nth-child": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		check, err := nthcheck.NthCheck(rule)
		if err != nil {
			return nil, err
		}

		if check.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}
		if check.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return helpers.GetElementParent(elem) != nil && next.Match(elem, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				siblings := domutils.GetSiblings(elem)
				pos := 0

				for i := 0; i < len(siblings); i++ {
					if elem == siblings[i] {
						break
					}
					if dom.IsTag(siblings[i]) {
						pos++
					}
				}

				return check.Fn(pos) && next.Match(elem, scope)
			},
		}, nil
	},
	"nth-last-child": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		check, err := nthcheck.NthCheck(rule)
		if err != nil {
			return nil, err
		}

		if check.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}
		if check.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return helpers.GetElementParent(elem) != nil && next.Match(elem, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				siblings := domutils.GetSiblings(elem)
				pos := 0

				for i := len(siblings) - 1; i >= 0; i-- {
					if elem == siblings[i] {
						break
					}
					if dom.IsTag(siblings[i]) {
						pos++
					}
				}

				return check.Fn(pos) && next.Match(elem, scope)
			},
		}, nil
	},
	"nth-of-type": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		check, err := nthcheck.NthCheck(rule)
		if err != nil {
			return nil, err
		}

		if check.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}
		if check.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return helpers.GetElementParent(elem) != nil && next.Match(elem, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				siblings := domutils.GetSiblings(elem)
				pos := 0

				for i := 0; i < len(siblings); i++ {
					currentSibling := siblings[i]
					if elem == currentSibling {
						break
					}
					if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == domutils.GetName(elem) {
						pos++
					}
				}

				return check.Fn(pos) && next.Match(elem, scope)
			},
		}, nil
	},
	"nth-last-of-type": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		check, err := nthcheck.NthCheck(rule)
		if err != nil {
			return nil, err
		}

		if check.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}
		if check.Type == types.MatchTypeAlwaysTrue {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return helpers.GetElementParent(elem) != nil && next.Match(elem, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				siblings := domutils.GetSiblings(elem)
				pos := 0

				for i := len(siblings) - 1; i >= 0; i-- {
					currentSibling := siblings[i]
					if elem == currentSibling {
						break
					}
					if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == domutils.GetName(elem) {
						pos++
					}
				}

				return check.Fn(pos) && next.Match(elem, scope)
			},
		}, nil
	},

	// TODO determine the actual root element
	"root": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return helpers.GetElementParent(elem) == nil && next.Match(elem, scope)
			},
		}, nil
	},

	"scope": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		if len(context) > 0 {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node, scope *dom.Node) bool {
					return slices.Index(context, elem) >= 0 && next.Match(elem, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				if scope != nil {
					return scope == elem && next.Match(elem, scope)
				}

				// Equivalent to :root
				return helpers.GetElementParent(elem) == nil && next.Match(elem, scope)
			},
		}, nil
	},
	// not supported
	"hover": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return false
			},
			Type: types.MatchTypeAlwaysFalse,
		}, nil
	},
	// not supported
	"visited": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return false
			},
			Type: types.MatchTypeAlwaysFalse,
		}, nil
	},
	// not supported
	"active": func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
	) (*types.CompiledQuery, error) {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return false
			},
			Type: types.MatchTypeAlwaysFalse,
		}, nil
	},
}
