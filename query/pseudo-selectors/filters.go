package pseudoselectors

import (
	"regexp"
	"slices"
	"strings"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/helpers"
	"github.com/krozhkov/go-css-select/query/internal"
	nthcheck "github.com/krozhkov/go-css-select/query/nth-check"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

/**
 * RFC 4647 extended filtering with pre-split subtags.
 * @param tag - Lowercased subtags of the element's language value.
 * @param rng - Lowercased subtags of the language range to match against.
 */
func extendedFilter(tag []string, rng []string) bool {
	hasTag := len(tag) > 0
	hasRng := len(rng) > 0
	if (!hasRng || rng[0] != "*") && (hasTag != hasRng || (hasTag && hasRng && rng[0] != tag[0])) {
		return false
	}

	tagIndex := 1

	for rangeIndex := 1; rangeIndex < len(rng); rangeIndex++ {
		if rng[rangeIndex] == "*" {
			continue
		}

		// Skip non-singleton tag subtags until we find a match.
		for tagIndex < len(tag) && tag[tagIndex] != rng[rangeIndex] {
			if len(tag[tagIndex]) <= 1 {
				return false
			}
			tagIndex++
		}

		if tagIndex >= len(tag) {
			return false
		}

		tagIndex++
	}

	return true
}

/** @see {@link https://www.w3.org/TR/selectors-4/#the-nth-child-pseudo} */
var nthOfRegex = regexp.MustCompile(`(?is)^(.+?)\s+of\s+(.+)$`)

/** A pre-compiled pseudo filter. */
type Filter = func(
	next *types.CompiledQuery,
	text string,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
) (*types.CompiledQuery, error)

func compileNth(reverse bool, ofType bool) Filter {
	return func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		var ofMatch []string
		if ofType {
			ofMatch = nthOfRegex.FindStringSubmatch(rule)
		}

		var nthRule string
		if len(ofMatch) > 0 {
			nthRule = strings.TrimSpace(ofMatch[1])
		} else {
			nthRule = rule
		}

		nthCheck, err := nthcheck.NthCheck(nthRule)
		if err != nil {
			return nil, err
		}

		if nthCheck.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}

		var ofSelector *types.CompiledQuery
		if len(ofMatch) > 0 && compileToken != nil {
			selector, err := parser.Parse(strings.TrimSpace(ofMatch[2]))
			if err != nil {
				return nil, err
			}

			ofSelector, err = compileToken(selector, helpers.CopyOptions(options), context)
			if err != nil {
				return nil, err
			}
		}

		if ofSelector != nil && ofSelector.Type == types.MatchTypeAlwaysFalse {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}, nil
		}

		if nthCheck.Type == types.MatchTypeAlwaysTrue && ofSelector == nil {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return helpers.GetElementParent(element) != nil && next.Match(element, scope)
				},
			}, nil
		}

		var shouldCount func(element *dom.Node, sibling *dom.Node, scope *dom.Node) bool
		if ofSelector != nil {
			shouldCount = func(_ *dom.Node, sibling *dom.Node, scope *dom.Node) bool {
				return ofSelector.Match(sibling, scope)
			}
		} else if ofType {
			shouldCount = func(element *dom.Node, sibling *dom.Node, _ *dom.Node) bool {
				return domutils.GetName(element) == domutils.GetName(sibling)
			}
		} else {
			shouldCount = func(_ *dom.Node, _ *dom.Node, _ *dom.Node) bool {
				return true // boolbase.trueFunc
			}
		}

		if reverse {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					if ofSelector != nil && !ofSelector.Match(element, scope) {
						return false
					}

					siblings := domutils.GetSiblings(element)
					pos := 0

					for index := len(siblings) - 1; index >= 0; index-- {
						sibling := siblings[index]
						if element == sibling {
							break
						}
						if dom.IsTag(sibling) && shouldCount(element, sibling, scope) {
							pos++
						}
					}

					return nthCheck.Fn(pos) && next.Match(element, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				if ofSelector != nil && !ofSelector.Match(element, scope) {
					return false
				}

				siblings := domutils.GetSiblings(element)
				pos := 0

				for _, sibling := range siblings {
					if element == sibling {
						break
					}
					if dom.IsTag(sibling) && shouldCount(element, sibling, scope) {
						pos++
					}
				}

				return nthCheck.Fn(pos) && next.Match(element, scope)
			},
		}, nil
	}
}

/**
 * Pre-compiled pseudo filters.
 */
var filters = map[string]Filter{
	"contains": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
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
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		itext := strings.ToLower(text)

		return helpers.CacheParentResults(next, options, func(elem *dom.Node, scope *dom.Node) bool {
			return strings.Contains(strings.ToLower(domutils.GetText(elem)), itext)
		}), nil
	},
	// Location specific methods
	"nth-child":        compileNth(false, false),
	"nth-last-child":   compileNth(true, false),
	"nth-of-type":      compileNth(false, true),
	"nth-last-of-type": compileNth(true, true),

	// TODO determine the actual root element
	"root": func(
		next *types.CompiledQuery,
		text string,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
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
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		if len(context) > 0 {
			return &types.CompiledQuery{
				Match: func(element *dom.Node, scope *dom.Node) bool {
					return slices.Index(context, element) >= 0 && next.Match(element, scope)
				},
			}, nil
		}

		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				if scope != nil {
					return scope == element && next.Match(element, scope)
				}

				// Equivalent to :root
				return helpers.GetElementParent(element) == nil && next.Match(element, scope)
			},
		}, nil
	},
	"lang": func(
		next *types.CompiledQuery,
		code string,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		ranges := toRanges(code)

		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				node := element

				for node != nil {
					var value *string
					if domutils.HasAttrib(node, "xml:lang") {
						value = new(domutils.GetAttributeValue(node, "xml:lang"))
					} else if domutils.HasAttrib(node, "lang") {
						value = new(domutils.GetAttributeValue(node, "lang"))
					}

					if value != nil {
						if *value == "" {
							return slices.ContainsFunc(
								ranges,
								func(r []string) bool {
									return len(r) > 0 && r[0] == ""
								},
							) && next.Match(element, scope)
						}

						tag := strings.Split(strings.ToLower(*value), "-")

						return slices.ContainsFunc(
							ranges,
							func(r []string) bool {
								return extendedFilter(tag, r)
							},
						) && next.Match(element, scope)
					}

					parent := domutils.GetParent(node)

					if parent != nil && dom.IsTag(parent) {
						node = parent
					} else {
						node = nil
					}
				}

				return slices.ContainsFunc(
					ranges,
					func(r []string) bool {
						return len(r) > 0 && r[0] == ""
					},
				) && next.Match(element, scope)
			},
		}, nil
	},
	"hover":   dynamicStatePseudo("isHovered"),
	"visited": dynamicStatePseudo("isVisited"),
	"active":  dynamicStatePseudo("isActive"),
}

/**
 * Dynamic state pseudos. These depend on optional Adapter methods.
 * @param name The name of the adapter method to call.
 * @returns Pseudo for the `filters` object.
 */
func dynamicStatePseudo(_ string) Filter {
	return func(
		next *types.CompiledQuery,
		rule string,
		options *types.Options,
		context []*dom.Node,
		compileToken types.CompileToken,
	) (*types.CompiledQuery, error) {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node, scope *dom.Node) bool {
				return false
			},
			Type: types.MatchTypeAlwaysFalse,
		}, nil
	}
}

func toRanges(code string) [][]string {
	ranges := internal.FilterFunc(
		internal.MapFunc(strings.Split(code, ","), strings.TrimSpace),
		func(r string) bool { return len(r) > 0 },
	)

	return internal.MapFunc(ranges, func(r string) []string {
		return strings.Split(strings.ToLower(stripEdgeQuotes(r)), "-")
	})
}

func stripEdgeQuotes(s string) string {
	if len(s) > 0 && (s[0] == '\'' || s[0] == '"') {
		s = s[1:]
	}
	if len(s) > 0 && (s[len(s)-1] == '\'' || s[len(s)-1] == '"') {
		s = s[:len(s)-1]
	}
	return s
}
