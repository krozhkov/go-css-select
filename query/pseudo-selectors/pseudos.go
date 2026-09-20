package pseudoselectors

import (
	"regexp"
	"slices"

	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

type Pseudo = func(
	element *dom.Node,
	options *types.Options,
) bool

/**
 * CSS limits the characters considered as whitespace to space, tab & line
 * feed. We add carriage returns as htmlparser2 doesn't normalize them to
 * line feeds.
 *
 * @see {@link https://www.w3.org/TR/css-text-3/#white-space}
 */
var isDocumentWhiteSpace = regexp.MustCompile(`^[ \t\r\n]*$`)

// While filters are precompiled, pseudos get called when they are needed
/** Runtime pseudo selector implementations. */
var pseudos = map[string]Pseudo{
	"empty": func(element *dom.Node, options *types.Options) bool {
		children := domutils.GetChildren(element)
		// First, make sure the tag does not have any element children.
		return !slices.ContainsFunc(children, dom.IsTag) &&
			// Then, check that the text content is only whitespace.
			internal.Every(children, func(elem *dom.Node) bool {
				// FIXME: `getText` call is potentially expensive.
				return isDocumentWhiteSpace.MatchString(domutils.GetText(elem))
			})
	},
	"first-child": func(element *dom.Node, _ *types.Options) bool {
		return domutils.PrevElementSibling(element) == nil
	},
	"last-child": func(element *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(element)

		for index := len(siblings) - 1; index >= 0; index-- {
			if element == siblings[index] {
				return true
			}
			if dom.IsTag(siblings[index]) {
				break
			}
		}

		return false
	},
	"first-of-type": func(element *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(element)
		elementName := domutils.GetName(element)

		for i := 0; i < len(siblings); i++ {
			currentSibling := siblings[i]
			if element == currentSibling {
				return true
			}
			if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == elementName {
				break
			}
		}

		return false
	},
	"last-of-type": func(element *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(element)
		elementName := domutils.GetName(element)

		for i := len(siblings) - 1; i >= 0; i-- {
			currentSibling := siblings[i]
			if element == currentSibling {
				return true
			}
			if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == elementName {
				break
			}
		}

		return false
	},
	"only-of-type": func(element *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(element)
		elementName := domutils.GetName(element)

		return internal.Every(siblings, func(sibling *dom.Node) bool {
			return element == sibling ||
				!dom.IsTag(sibling) ||
				domutils.GetName(sibling) != elementName
		})
	},
	"only-child": func(element *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(element)

		return internal.Every(siblings, func(sibling *dom.Node) bool {
			return element == sibling || !dom.IsTag(sibling)
		})
	},
}
