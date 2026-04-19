package parser

import (
	"strings"
)

func charsToEscapeInAttributeValue(b byte) bool {
	switch b {
	case '\\', '"':
		return true
	default:
		return false
	}
}

func charsToEscapeInPseudoValue(b byte) bool {
	switch b {
	case '\\', '"', '(', ')':
		return true
	default:
		return false
	}
}

func charsToEscapeInName(b byte) bool {
	switch b {
	case '\\',
		'"',
		'(',
		')',
		'~',
		'^',
		'$',
		'*',
		'+',
		'!',
		'|',
		':',
		'[',
		']',
		' ',
		'.',
		'%':
		return true
	default:
		return false
	}
}

/**
 * Turns `selector` back into a string.
 *
 * @param selector Selector to stringify.
 */
func Stringify(selector [][]*Selector) string {
	var sb = new(strings.Builder)
	stringify(sb, selector)
	return sb.String()
}

func stringify(sb *strings.Builder, selector [][]*Selector) {
	lastIndex := len(selector) - 1
	for i, tokens := range selector {
		for index, token := range tokens {
			stringifyToken(sb, token, index, tokens)
		}

		if i != lastIndex {
			sb.WriteString(", ")
		}
	}
}

func stringifyToken(sb *strings.Builder, token *Selector, index int, array []*Selector) {
	switch token.Type {
	// Simple types
	case SelectorTypeChild:
		{
			if index == 0 {
				sb.WriteString("> ")
			} else {
				sb.WriteString(" > ")
			}
		}
	case SelectorTypeParent:
		{
			if index == 0 {
				sb.WriteString("< ")
			} else {
				sb.WriteString(" < ")
			}
		}
	case SelectorTypeSibling:
		{
			if index == 0 {
				sb.WriteString("~ ")
			} else {
				sb.WriteString(" ~ ")
			}
		}
	case SelectorTypeAdjacent:
		{
			if index == 0 {
				sb.WriteString("+ ")
			} else {
				sb.WriteString(" + ")
			}
		}
	case SelectorTypeDescendant:
		{
			sb.WriteString(" ")
		}
	case SelectorTypeColumnCombinator:
		{
			if index == 0 {
				sb.WriteString("|| ")
			} else {
				sb.WriteString(" || ")
			}
		}
	case SelectorTypeUniversal:
		{
			// Return an empty string if the selector isn't needed.
			if token.Namespace != nil && *token.Namespace == "*" && index+1 < len(array) && (array[index+1].Type == SelectorTypeUniversal || IsTraversal(array[index+1])) {
				// do nothing
			} else {
				getNamespace(sb, token.Namespace)
				sb.WriteRune('*')
			}
		}
	case SelectorTypeTag:
		{
			getNamespacedName(sb, token)
		}
	case SelectorTypePseudoElement:
		{
			sb.WriteString("::")
			escapeName(sb, token.Name, charsToEscapeInName)
			if token.Data != nil {
				sb.WriteRune('(')
				escapeName(sb, *token.Data, charsToEscapeInPseudoValue)
				sb.WriteRune(')')
			}
		}
	case SelectorTypePseudo:
		{
			sb.WriteString(":")
			escapeName(sb, token.Name, charsToEscapeInName)
			if token.Data != nil {
				sb.WriteRune('(')
				escapeName(sb, *token.Data, charsToEscapeInPseudoValue)
				sb.WriteRune(')')
			}
			if len(token.Children) > 0 {
				sb.WriteRune('(')
				stringify(sb, token.Children)
				sb.WriteRune(')')
			}
		}
	case SelectorTypeAttribute:
		{
			if token.Name == "id" &&
				token.Action == AttributeActionEquals &&
				token.IgnoreCase == IgnoreCaseModeQuirksMode &&
				token.Namespace == nil {
				sb.WriteRune('#')
				escapeName(sb, *token.Data, charsToEscapeInName)
				return
			}

			if token.Name == "class" &&
				token.Action == AttributeActionElement &&
				token.IgnoreCase == IgnoreCaseModeQuirksMode &&
				token.Namespace == nil {
				sb.WriteRune('.')
				escapeName(sb, *token.Data, charsToEscapeInName)
				return
			}

			//const name = getNamespacedName(token);

			if token.Action == AttributeActionExists {
				sb.WriteRune('[')
				getNamespacedName(sb, token)
				sb.WriteRune(']')
				return
			}

			sb.WriteRune('[')
			getNamespacedName(sb, token)
			sb.WriteString(getActionValue(token.Action))
			sb.WriteString("=\"")
			escapeName(sb, *token.Data, charsToEscapeInAttributeValue)
			sb.WriteString("\"")

			switch token.IgnoreCase {
			case IgnoreCaseModeUnknown:
				// do nothing
			case IgnoreCaseModeIgnoreCase:
				sb.WriteString(" i")
			default:
				sb.WriteString(" s")
			}

			sb.WriteRune(']')
		}
	}
}

func getActionValue(action AttributeAction) string {
	switch action {
	case AttributeActionEquals:
		return ""
	case AttributeActionElement:
		return "~"
	case AttributeActionStart:
		return "^"
	case AttributeActionEnd:
		return "$"
	case AttributeActionAny:
		return "*"
	case AttributeActionNot:
		return "!"
	case AttributeActionHyphen:
		return "|"
	default:
		return ""
	}
}

func getNamespacedName(sb *strings.Builder, token *Selector) {
	getNamespace(sb, token.Namespace)
	escapeName(sb, token.Name, charsToEscapeInName)
}

func getNamespace(sb *strings.Builder, namespace *string) {
	if namespace != nil {
		if *namespace == "*" {
			sb.WriteRune('*')
		} else {
			escapeName(sb, *namespace, charsToEscapeInName)
		}
		sb.WriteRune('|')
	}
}

func escapeName(sb *strings.Builder, name string, isCharsToEscape func(b byte) bool) {
	lastIndex := 0
	length := sb.Len()

	for index := 0; index < len(name); index++ {
		if isCharsToEscape(name[index]) {
			sb.WriteString(name[lastIndex:index])
			sb.WriteRune('\\')
			sb.WriteByte(name[index])
			lastIndex = index + 1
		}
	}

	if sb.Len() > length {
		if lastIndex < len(name) {
			sb.WriteString(name[lastIndex:])
		}
	} else {
		sb.WriteString(name)
	}
}
