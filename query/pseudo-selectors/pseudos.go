package pseudoselectors

import (
	"regexp"

	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

type Pseudo = func(
	elem *dom.Node,
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
var pseudos = map[string]Pseudo{
	"empty": func(elem *dom.Node, options *types.Options) bool {
		children := domutils.GetChildren(elem)
		// First, make sure the tag does not have any element children.
		return internal.Every(children, dom.IsTag) &&
			// Then, check that the text content is only whitespace.
			internal.Every(children, func(elem *dom.Node) bool {
				// FIXME: `getText` call is potentially expensive.
				return isDocumentWhiteSpace.MatchString(domutils.GetText(elem))
			})
	},
	"first-child": func(elem *dom.Node, _ *types.Options) bool {
		return domutils.PrevElementSibling(elem) == nil
	},
	"last-child": func(elem *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(elem)

		for i := len(siblings) - 1; i >= 0; i-- {
			if elem == siblings[i] {
				return true
			}
			if dom.IsTag(siblings[i]) {
				break
			}
		}

		return false
	},
	"first-of-type": func(elem *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(elem)
		elemName := domutils.GetName(elem)

		for i := 0; i < len(siblings); i++ {
			currentSibling := siblings[i]
			if elem == currentSibling {
				return true
			}
			if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == elemName {
				break
			}
		}

		return false
	},
	"last-of-type": func(elem *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(elem)
		elemName := domutils.GetName(elem)

		for i := len(siblings) - 1; i >= 0; i-- {
			currentSibling := siblings[i]
			if elem == currentSibling {
				return true
			}
			if dom.IsTag(currentSibling) && domutils.GetName(currentSibling) == elemName {
				break
			}
		}

		return false
	},
	"only-of-type": func(elem *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(elem)
		elemName := domutils.GetName(elem)

		return internal.Every(siblings, func(sibling *dom.Node) bool {
			return elem == sibling ||
				!dom.IsTag(sibling) ||
				domutils.GetName(sibling) != elemName
		})
	},
	"only-child": func(elem *dom.Node, options *types.Options) bool {
		siblings := domutils.GetSiblings(elem)

		return internal.Every(siblings, func(sibling *dom.Node) bool {
			return elem == sibling || !dom.IsTag(sibling)
		})
	},
}
