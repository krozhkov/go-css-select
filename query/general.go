package query

import (
	"errors"
	"slices"
	"strings"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	pseudoselectors "github.com/krozhkov/go-css-select/query/pseudo-selectors"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

func ptr[T any](v T) *T {
	return &v
}

/*
 * All available rules
 */
func CompileGeneralSelector(
	next *types.CompiledQuery,
	selector *parser.Selector,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
	hasExpensiveSubselector bool,
) (*types.CompiledQuery, error) {
	switch selector.Type {
	case parser.SelectorTypePseudoElement:
		return nil, errors.New("pseudo-elements are not supported by css-select")
	case parser.SelectorTypeColumnCombinator:
		return nil, errors.New("column combinators are not yet supported by css-select")
	case parser.SelectorTypeAttribute:
		{
			if selector.Namespace != nil {
				return nil, errors.New("namespaced attributes are not yet supported by css-select")
			}

			if options == nil || options.XmlMode != types.OptYes || options.LowerCaseAttributeNames == types.OptYes {
				selector.Name = strings.ToLower(selector.Name)
			}
			return attributeRules[selector.Action](next, selector, options), nil
		}
	case parser.SelectorTypePseudo:
		{
			return pseudoselectors.CompilePseudoSelector(
				next,
				selector,
				options,
				context,
				compileToken,
			)
		}
	// Tags
	case parser.SelectorTypeTag:
		{
			if selector.Namespace != nil {
				return nil, errors.New("namespaced tag names are not yet supported by css-select")
			}

			name := selector.Name

			if options == nil || options.XmlMode != types.OptYes || options.LowerCaseTags == types.OptYes {
				name = strings.ToLower(name)
			}

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return domutils.GetName(elem) == name && next.Match(elem)
				},
			}, nil
		}
	// Traversal
	case parser.SelectorTypeDescendant:
		{
			if !hasExpensiveSubselector || options.CacheResults == types.OptNo {
				return &types.CompiledQuery{
					Match: func(elem *dom.Node) bool {
						for current := helpers.GetElementParent(elem); current != nil; current = helpers.GetElementParent(current) {
							if next.Match(current) {
								return true
							}
						}

						return false
					},
				}, nil
			}

			resultCache := helpers.NewCache[dom.Node, *bool]()

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					var result *bool

					for current := helpers.GetElementParent(elem); current != nil; current = helpers.GetElementParent(current) {
						cached := resultCache.Get(current)

						if cached == nil {
							result = ptr(next.Match(current))
							resultCache.Set(current, result)
							if *result {
								return true
							}
						} else {
							if result != nil {
								result = cached
							}
							return *cached
						}
					}

					return false
				},
			}, nil
		}
	case "_flexibleDescendant":
		{
			// Include element itself, only used while querying an array
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					current := elem

					for {
						if next.Match(current) {
							return true
						}
						current = helpers.GetElementParent(current)
						if current == nil {
							break
						}
					}

					return false
				},
			}, nil
		}
	case parser.SelectorTypeParent:
		{
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					children := domutils.GetChildren(elem)
					return slices.IndexFunc(children, func(n *dom.Node) bool {
						return dom.IsTag(n) && next.Match(n)
					}) >= 0
				},
			}, nil
		}
	case parser.SelectorTypeChild:
		{
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					parent := helpers.GetElementParent(elem)
					return parent != nil && next.Match(parent)
				},
			}, nil
		}
	case parser.SelectorTypeSibling:
		{
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					siblings := domutils.GetSiblings(elem)

					for i := 0; i < len(siblings); i++ {
						currentSibling := siblings[i]
						if elem == currentSibling {
							break
						}
						if dom.IsTag(currentSibling) && next.Match(currentSibling) {
							return true
						}
					}

					return false
				},
			}, nil
		}
	case parser.SelectorTypeAdjacent:
		{
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					previous := domutils.PrevElementSibling(elem)
					return previous != nil && next.Match(previous)
				},
			}, nil
		}
	case parser.SelectorTypeUniversal:
		{
			if selector.Namespace != nil && *selector.Namespace != "*" {
				return nil, errors.New("namespaced universal selectors are not yet supported by css-select")
			}

			return next, nil
		}
	default:
		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return false
			},
		}, nil
	}
}
