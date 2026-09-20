package helpers

import (
	"slices"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/internal"
)

/**
 * Check whether a selector token performs traversal.
 * @param token Selector token(s) to compile.
 */
func IsTraversal(token *parser.Selector) bool {
	return token.Type == "_flexibleDescendant" || parser.IsTraversal(token)
}

/**
 * Sort the parts of the passed selector, as there is potential for
 * optimization (some types of selectors are faster than others).
 * @param arr Selector to sort
 */
func SortRules(array []*parser.Selector) {
	ratings := internal.MapFunc(array, GetQuality)
	for index := 1; index < len(array); index++ {
		procNew := ratings[index]

		if procNew < 0 {
			continue
		}

		// Use insertion sort to move the token to the correct position.
		for currentIndex := index; currentIndex > 0 && procNew < ratings[currentIndex-1]; currentIndex-- {
			token := array[currentIndex]
			array[currentIndex] = array[currentIndex-1]
			array[currentIndex-1] = token
			ratings[currentIndex] = ratings[currentIndex-1]
			ratings[currentIndex-1] = procNew
		}
	}
}

func getAttributeQuality(token *parser.Selector) int {
	switch token.Action {
	case parser.AttributeActionExists:
		return 10
	case parser.AttributeActionEquals:
		// Prefer ID selectors (eg. #ID)
		if token.Name == "id" {
			return 9
		} else {
			return 8
		}
	case parser.AttributeActionNot:
		return 7
	case parser.AttributeActionStart:
		return 6
	case parser.AttributeActionEnd:
		return 6
	case parser.AttributeActionAny:
		return 5
	case parser.AttributeActionHyphen:
		return 4
	case parser.AttributeActionElement:
		return 3
	default:
		return -1
	}
}

/**
 * Determine the quality of the passed token. The higher the number, the
 * faster the token is to execute.
 * @param token Token to get the quality of.
 * @returns The token's quality.
 */
func GetQuality(token *parser.Selector) int {
	switch token.Type {
	case parser.SelectorTypeUniversal:
		{
			return 50
		}
	case parser.SelectorTypeTag:
		{
			return 30
		}
	case parser.SelectorTypeAttribute:
		{
			var ignoreCaseFactor = 1
			if token.IgnoreCase == parser.IgnoreCaseModeQuirksMode || token.IgnoreCase == parser.IgnoreCaseModeIgnoreCase {
				// `ignoreCase` adds some overhead, half the result if applicable.
				ignoreCaseFactor = 2
			}
			return getAttributeQuality(token) / ignoreCaseFactor
		}
	case parser.SelectorTypePseudo:
		{
			if (token.Data != nil && *token.Data != "") || len(token.Children) > 0 {
				if token.Name == "has" || token.Name == "contains" || token.Name == "icontains" {
					// Expensive in any case — run as late as possible.
					return 0
				}

				if len(token.Children) > 0 {
					// Eg. `:is`, `:not`
					quality := slices.Min(internal.MapFunc(token.Children, func(d []*parser.Selector) int {
						return slices.Min(internal.MapFunc(d, GetQuality))
					}))
					if quality < 0 {
						// If we have traversals, try to avoid executing this selector
						return 0
					}
					return quality
				}

				return 2
			} else {
				return 3
			}
		}
	default:
		{
			return -1
		}
	}
}

/**
 * Check whether a token or nested token includes `:scope`.
 * @param t Selector token under inspection.
 */
func IncludesScopePseudo(t *parser.Selector) bool {
	return t.Type == parser.SelectorTypePseudo &&
		(t.Name == "scope" || (len(t.Children) > 0 && slices.IndexFunc(t.Children, func(d []*parser.Selector) bool { return slices.IndexFunc(d, IncludesScopePseudo) >= 0 }) >= 0))
}
