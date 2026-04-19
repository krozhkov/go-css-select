package query

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

var spaceRe = regexp.MustCompile(`\s`)

/**
 * Attributes that are case-insensitive in HTML.
 *
 * @private
 * @see https://html.spec.whatwg.org/multipage/semantics-other.html#case-sensitivity-of-selectors
 */
func isCaseInsensitiveAttribute(attr string) bool {
	switch attr {
	case "accept",
		"accept-charset",
		"align",
		"alink",
		"axis",
		"bgcolor",
		"charset",
		"checked",
		"clear",
		"codetype",
		"color",
		"compact",
		"declare",
		"defer",
		"dir",
		"direction",
		"disabled",
		"enctype",
		"face",
		"frame",
		"hreflang",
		"http-equiv",
		"lang",
		"language",
		"link",
		"media",
		"method",
		"multiple",
		"nohref",
		"noresize",
		"noshade",
		"nowrap",
		"readonly",
		"rel",
		"rev",
		"rules",
		"scope",
		"scrolling",
		"selected",
		"shape",
		"target",
		"text",
		"type",
		"valign",
		"valuetype",
		"vlink":
		return true
	default:
		return false
	}
}

func shouldIgnoreCase(
	selector *parser.Selector,
	options *types.Options,
) bool {
	switch selector.IgnoreCase {
	case parser.IgnoreCaseModeCaseSensitive:
		return false
	case parser.IgnoreCaseModeIgnoreCase:
		return true
	case parser.IgnoreCaseModeQuirksMode:
		return options != nil && options.QuirksMode == types.OptYes
	default:
		return (options == nil || options.XmlMode != types.OptYes) && isCaseInsensitiveAttribute(selector.Name)
	}
}

type Attribute = func(
	next *types.CompiledQuery,
	selector *parser.Selector,
	options *types.Options,
) *types.CompiledQuery

/**
 * Attribute selectors
 */
var attributeRules = map[parser.AttributeAction]Attribute{
	"equals": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}

		if shouldIgnoreCase(selector, options) {
			value = strings.ToLower(value)

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" &&
						len(attr) == len(value) &&
						strings.ToLower(attr) == value &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return domutils.GetAttributeValue(elem, name) == value && next.Match(elem)
			},
		}
	},
	"hyphen": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}
		length := len(value)

		if shouldIgnoreCase(selector, options) {
			value = strings.ToLower(value)

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" &&
						(len(attr) == length || attr[length] == '-') &&
						strings.ToLower(attr[:length]) == value &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return attr != "" &&
					len(attr) >= length &&
					(len(attr) == length || attr[length] == '-') &&
					attr[:length] == value &&
					next.Match(elem)
			},
		}
	},
	"element": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}
		if spaceRe.MatchString(value) {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}
		}

		flags := ""
		if shouldIgnoreCase(selector, options) {
			flags = "(?i)"
		}

		regex := regexp.MustCompile(fmt.Sprintf(`%s(?:^|\s)%s(?:$|\s)`, flags, regexp.QuoteMeta(value)))

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return attr != "" &&
					len(attr) >= len(value) &&
					regex.MatchString(attr) &&
					next.Match(elem)
			},
		}
	},
	"exists": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return domutils.HasAttrib(elem, selector.Name) && next.Match(elem)
			},
		}
	},
	"start": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}
		length := len(value)

		if length == 0 {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}
		}

		if shouldIgnoreCase(selector, options) {
			value = strings.ToLower(value)

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" &&
						len(attr) >= length &&
						strings.ToLower(attr[:length]) == value &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return attr != "" &&
					strings.HasPrefix(attr, value) &&
					next.Match(elem)
			},
		}
	},
	"end": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}
		length := len(value)

		if length == 0 {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}
		}

		if shouldIgnoreCase(selector, options) {
			value = strings.ToLower(value)

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" &&
						len(attr) >= length &&
						strings.ToLower(attr[len(attr)-length:]) == value &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return attr != "" &&
					strings.HasSuffix(attr, value) &&
					next.Match(elem)
			},
		}
	},
	"any": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}

		if value == "" {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					return false
				},
				Type: types.MatchTypeAlwaysFalse,
			}
		}

		if shouldIgnoreCase(selector, options) {
			regex := regexp.MustCompile(fmt.Sprintf(`(?i)%s`, regexp.QuoteMeta(value)))

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" &&
						len(attr) >= len(value) &&
						regex.MatchString(attr) &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return attr != "" &&
					strings.Contains(attr, value) &&
					next.Match(elem)
			},
		}
	},
	"not": func(next *types.CompiledQuery, selector *parser.Selector, options *types.Options) *types.CompiledQuery {
		name := selector.Name
		var value string
		if selector.Data != nil {
			value = *selector.Data
		}

		if value == "" {
			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return attr != "" && next.Match(elem)
				},
			}
		}

		if shouldIgnoreCase(selector, options) {
			value = strings.ToLower(value)

			return &types.CompiledQuery{
				Match: func(elem *dom.Node) bool {
					attr := domutils.GetAttributeValue(elem, name)
					return (attr == "" || len(attr) != len(value) || strings.ToLower(attr) != value) &&
						next.Match(elem)
				},
			}
		}

		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				attr := domutils.GetAttributeValue(elem, name)
				return (attr == "" || attr != value) && next.Match(elem)
			},
		}
	},
}
