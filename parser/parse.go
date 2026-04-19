package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reName = regexp.MustCompile(`^[^#\\]?(?:\\(?:[0-9a-fA-F]{1,6}\s?|.)|[\w\x{00B0}-\x{10FFFF}-])+`)
var reEscape = regexp.MustCompile(`(?i)\\([0-9a-f]{1,6}\s?|(\s)|.)`)

const (
	LeftParenthesis    = 40
	RightParenthesis   = 41
	LeftSquareBracket  = 91
	RightSquareBracket = 93
	Comma              = 44
	Period             = 46
	Colon              = 58
	SingleQuote        = 39
	DoubleQuote        = 34
	Plus               = 43
	Tilde              = 126
	QuestionMark       = 63
	ExclamationMark    = 33
	Slash              = 47
	Equal              = 61
	Dollar             = 36
	Pipe               = 124
	Circumflex         = 94
	Asterisk           = 42
	GreaterThan        = 62
	LessThan           = 60
	Hash               = 35
	LowerI             = 105
	LowerS             = 115
	BackSlash          = 92

	// Whitespace
	Space          = 32
	Tab            = 9
	NewLine        = 10
	FormFeed       = 12
	CarriageReturn = 13
)

var actionTypes = map[byte]AttributeAction{
	Tilde:           AttributeActionElement,
	Circumflex:      AttributeActionStart,
	Dollar:          AttributeActionEnd,
	Asterisk:        AttributeActionAny,
	ExclamationMark: AttributeActionNot,
	Pipe:            AttributeActionHyphen,
}

// Pseudos, whose data property is parsed as well.
func isUnpackPseudos(name string) bool {
	switch name {
	case "has",
		"not",
		"matches",
		"is",
		"where",
		"host",
		"host-context":
		return true
	default:
		return false
	}
}

/**
 * Pseudo elements defined in CSS Level 1 and CSS Level 2 can be written with
 * a single colon; eg. :before will turn into ::before.
 *
 * @see {@link https://www.w3.org/TR/2018/WD-selectors-4-20181121/#pseudo-element-syntax}
 */
func isPseudosToPseudoElements(name string) bool {
	switch name {
	case "before",
		"after",
		"first-line",
		"first-letter":
		return true
	default:
		return false
	}
}

func String(s string) *string {
	return &s
}

/**
 * Checks whether a specific selector is a traversal.
 * This is useful eg. in swapping the order of elements that
 * are not traversals.
 *
 * @param selector Selector to check.
 */
func IsTraversal(selector *Selector) bool {
	switch selector.Type {
	case SelectorTypeAdjacent, SelectorTypeChild, SelectorTypeDescendant, SelectorTypeParent, SelectorTypeSibling, SelectorTypeColumnCombinator:
		return true
	default:
		return false
	}
}

func isStripQuotesFromPseudos(name string) bool {
	switch name {
	case "contains", "icontains":
		return true
	default:
		return false
	}
}

// Unescape function taken from https://github.com/jquery/sizzle/blob/master/src/sizzle.js#L152
func funescape(escaped string) string {
	codePoint, err := strconv.ParseInt(strings.TrimSpace(escaped), 16, 32)

	// NaN means non-codepoint
	if err != nil || escaped == "" {
		return escaped
	}

	if codePoint == 0 {
		return string(rune(0xFFFD))
	}

	return string(rune(codePoint))
}

func unescapeCSS(cssString string) string {
	return reEscape.ReplaceAllStringFunc(cssString, func(escaped string) string {
		return funescape(escaped[1:])
	})
}

func isQuote(c byte) bool {
	return c == SingleQuote || c == DoubleQuote
}

func isWhitespace(c byte) bool {
	return c == Space ||
		c == Tab ||
		c == NewLine ||
		c == FormFeed ||
		c == CarriageReturn
}

/**
 * Parses `selector`.
 *
 * @param selector Selector to parse.
 * @returns Returns a two-dimensional array.
 * The first dimension represents selectors separated by commas (eg. `sub1, sub2`),
 * the second contains the relevant tokens for that selector.
 */
func Parse(selector string) ([][]*Selector, error) {
	subselects := make([][]*Selector, 0)

	endIndex, err := parseSelector(&subselects, selector, 0)

	if err != nil {
		return nil, err
	}

	if endIndex < len(selector) {
		return nil, fmt.Errorf("Unmatched selector: %s", selector[endIndex:])
	}

	return subselects, nil
}

func parseSelector(
	subselects *[][]*Selector,
	selector string,
	selectorIndex int,
) (int, error) {
	tokens := make([]*Selector, 0)

	getName := func(offset int) (string, error) {
		match := reName.FindStringSubmatch(selector[selectorIndex+offset:])

		if len(match) == 0 {
			return "", fmt.Errorf("Expected name, found  %s", selector[selectorIndex:])
		}

		name := match[0]
		selectorIndex += offset + len(name)
		return unescapeCSS(name), nil
	}

	stripWhitespace := func(offset int) {
		selectorIndex += offset

		for selectorIndex < len(selector) && isWhitespace(selector[selectorIndex]) {
			selectorIndex++
		}
	}

	readValueWithParenthesis := func() (string, error) {
		selectorIndex += 1
		start := selectorIndex

		for counter := 1; selectorIndex < len(selector); selectorIndex++ {
			switch selector[selectorIndex] {
			case BackSlash:
				{
					// Skip next character
					selectorIndex += 1
					break
				}
			case LeftParenthesis:
				{
					counter += 1
					break
				}
			case RightParenthesis:
				{
					counter -= 1

					if counter == 0 {
						value := unescapeCSS(selector[start:selectorIndex])
						selectorIndex++
						return value, nil
					}

					break
				}
			}
		}

		return "", errors.New("parenthesis not matched")
	}

	ensureNotTraversal := func() error {
		if len(tokens) > 0 && IsTraversal(tokens[len(tokens)-1]) {
			return errors.New("did not expect successive traversals.")
		}

		return nil
	}

	addTraversal := func(typ SelectorType) error {
		if len(tokens) > 0 && tokens[len(tokens)-1].Type == SelectorTypeDescendant {
			tokens[len(tokens)-1].Type = typ
			return nil
		}

		err := ensureNotTraversal()
		if err != nil {
			return err
		}

		tokens = append(tokens, &Selector{Type: typ})

		return nil
	}

	addSpecialAttribute := func(name string, action AttributeAction) error {
		value, err := getName(1)
		if err != nil {
			return err
		}

		tokens = append(tokens, &Selector{
			Type:       SelectorTypeAttribute,
			Name:       name,
			Action:     action,
			Data:       String(value),
			IgnoreCase: IgnoreCaseModeQuirksMode,
		})

		return nil
	}

	/**
	 * We have finished parsing the current part of the selector.
	 *
	 * Remove descendant tokens at the end if they exist,
	 * and return the last index, so that parsing can be
	 * picked up from here.
	 */
	finalizeSubselector := func() error {
		if len(tokens) > 0 && tokens[len(tokens)-1].Type == SelectorTypeDescendant {
			lastIndex := len(tokens) - 1
			tokens[lastIndex] = nil
			tokens = tokens[:lastIndex]
		}

		if len(tokens) == 0 {
			return errors.New("empty sub-selector")
		}

		*subselects = append(*subselects, tokens)

		return nil
	}

	stripWhitespace(0)

	if len(selector) == selectorIndex {
		return selectorIndex, nil
	}

loop:
	for selectorIndex < len(selector) {
		firstChar := selector[selectorIndex]

		switch firstChar {
		// Whitespace
		case Space, Tab, NewLine, FormFeed, CarriageReturn:
			{
				if len(tokens) == 0 || tokens[0].Type != SelectorTypeDescendant {
					err := ensureNotTraversal()
					if err != nil {
						return selectorIndex, err
					}
					tokens = append(tokens, &Selector{Type: SelectorTypeDescendant})
				}

				stripWhitespace(1)
				break
			}
		// Traversals
		case GreaterThan:
			{
				err := addTraversal(SelectorTypeChild)
				if err != nil {
					return selectorIndex, err
				}
				stripWhitespace(1)
				break
			}
		case LessThan:
			{
				err := addTraversal(SelectorTypeParent)
				if err != nil {
					return selectorIndex, err
				}
				stripWhitespace(1)
				break
			}
		case Tilde:
			{
				err := addTraversal(SelectorTypeSibling)
				if err != nil {
					return selectorIndex, err
				}
				stripWhitespace(1)
				break
			}
		case Plus:
			{
				err := addTraversal(SelectorTypeAdjacent)
				if err != nil {
					return selectorIndex, err
				}
				stripWhitespace(1)
				break
			}
		// Special attribute selectors: .class, #id
		case Period:
			{
				err := addSpecialAttribute("class", AttributeActionElement)
				if err != nil {
					return selectorIndex, err
				}
				break
			}
		case Hash:
			{
				err := addSpecialAttribute("id", AttributeActionEquals)
				if err != nil {
					return selectorIndex, err
				}
				break
			}
		case LeftSquareBracket:
			{
				stripWhitespace(1)

				// Determine attribute name and namespace

				var name string
				var namespace *string
				var err error

				if selectorIndex < len(selector) && selector[selectorIndex] == Pipe {
					// Equivalent to no namespace
					name, err = getName(1)
					if err != nil {
						return selectorIndex, err
					}
				} else if strings.HasPrefix(selector[selectorIndex:], "*|") {
					namespace = String("*")
					name, err = getName(2)
					if err != nil {
						return selectorIndex, err
					}
				} else {
					name, err = getName(0)
					if err != nil {
						return selectorIndex, err
					}

					if selector[selectorIndex] == Pipe && selector[selectorIndex+1] != Equal {
						namespace = String(name)
						name, err = getName(1)
						if err != nil {
							return selectorIndex, err
						}
					}
				}

				stripWhitespace(0)

				// Determine comparison operation

				action := AttributeActionExists
				possibleAction, ok := actionTypes[selector[selectorIndex]]

				if ok {
					action = possibleAction

					if selectorIndex+1 >= len(selector) || selector[selectorIndex+1] != Equal {
						return selectorIndex, errors.New("expected `=`")
					}

					stripWhitespace(2)
				} else if selector[selectorIndex] == Equal {
					action = AttributeActionEquals
					stripWhitespace(1)
				}

				// Determine value

				var value = String("")
				var ignoreCase IgnoreCaseMode = IgnoreCaseModeUnknown

				if action != "exists" {
					if isQuote(selector[selectorIndex]) {
						quote := selector[selectorIndex]
						selectorIndex += 1
						sectionStart := selectorIndex
						for selectorIndex < len(selector) && selector[selectorIndex] != quote {
							// Skip next character if it is escaped
							if selector[selectorIndex] == BackSlash {
								selectorIndex += 2
							} else {
								selectorIndex += 1
							}
						}

						if selectorIndex >= len(selector) || selector[selectorIndex] != quote {
							return selectorIndex, errors.New("attribute value didn't end")
						}

						value = String(unescapeCSS(selector[sectionStart:selectorIndex]))
						selectorIndex += 1
					} else {
						valueStart := selectorIndex

						for selectorIndex < len(selector) && !isWhitespace(selector[selectorIndex]) && selector[selectorIndex] != RightSquareBracket {
							// Skip next character if it is escaped
							if selector[selectorIndex] == BackSlash {
								selectorIndex += 2
							} else {
								selectorIndex += 1
							}
						}

						value = String(unescapeCSS(selector[valueStart:selectorIndex]))
					}

					stripWhitespace(0)

					if selectorIndex < len(selector) {
						// See if we have a force ignore flag
						switch selector[selectorIndex] | 0x20 {
						// If the forceIgnore flag is set (either `i` or `s`), use that value
						case LowerI:
							{
								ignoreCase = IgnoreCaseModeIgnoreCase
								stripWhitespace(1)
								break
							}
						case LowerS:
							{
								ignoreCase = IgnoreCaseModeCaseSensitive
								stripWhitespace(1)
								break
							}
						}
					}
				}

				if selectorIndex >= len(selector) || selector[selectorIndex] != RightSquareBracket {
					return selectorIndex, errors.New("attribute selector didn't terminate")
				}

				selectorIndex += 1

				attributeSelector := &Selector{
					Type:       SelectorTypeAttribute,
					Name:       name,
					Action:     action,
					Data:       value,
					Namespace:  namespace,
					IgnoreCase: ignoreCase,
				}

				tokens = append(tokens, attributeSelector)
				break
			}
		case Colon:
			{
				if selectorIndex+1 < len(selector) && selector[selectorIndex+1] == Colon {
					name, err := getName(2)
					if err != nil {
						return selectorIndex, err
					}

					var value *string
					if selectorIndex < len(selector) && selector[selectorIndex] == LeftParenthesis {
						data, err := readValueWithParenthesis()
						if err != nil {
							return selectorIndex, err
						}
						value = String(data)
					}

					tokens = append(tokens, &Selector{
						Type: SelectorTypePseudoElement,
						Name: strings.ToLower(name),
						Data: value,
					})
					break
				}

				name, err := getName(1)
				if err != nil {
					return selectorIndex, err
				}
				name = strings.ToLower(name)

				if isPseudosToPseudoElements(name) {
					tokens = append(tokens, &Selector{
						Type: SelectorTypePseudoElement,
						Name: name,
					})
					break
				}

				var value *string
				var children [][]*Selector

				if selectorIndex < len(selector) && selector[selectorIndex] == LeftParenthesis {
					if isUnpackPseudos(name) {
						if selectorIndex+1 >= len(selector) || isQuote(selector[selectorIndex+1]) {
							return selectorIndex, errors.New(`pseudo-selector ${name} cannot be quoted`)
						}

						children = make([][]*Selector, 0)
						selectorIndex, err = parseSelector(
							&children,
							selector,
							selectorIndex+1,
						)

						if err != nil {
							return selectorIndex, err
						}

						if selectorIndex >= len(selector) || selector[selectorIndex] != RightParenthesis {
							return selectorIndex, errors.New(`missing closing parenthesis in :${name} (${selector})`)
						}

						selectorIndex += 1
					} else {
						data, err := readValueWithParenthesis()
						if err != nil {
							return selectorIndex, err
						}

						if isStripQuotesFromPseudos(name) && len(data) > 0 {
							quot := data[0]

							if quot == data[len(data)-1] && isQuote(quot) {
								data = data[1 : len(data)-1]
							}
						}

						value = String(unescapeCSS(data))
					}
				}

				tokens = append(tokens, &Selector{
					Type:     SelectorTypePseudo,
					Name:     name,
					Data:     value,
					Children: children,
				})
				break
			}
		case Comma:
			{
				err := finalizeSubselector()
				if err != nil {
					return selectorIndex, err
				}

				tokens = make([]*Selector, 0)
				stripWhitespace(1)
				break
			}
		default:
			{
				if strings.HasPrefix(selector[selectorIndex:], "/*") {
					endIndex := strings.Index(selector[selectorIndex+2:], "*/")
					if endIndex != -1 {
						endIndex += selectorIndex + 2
					}

					if endIndex == -1 {
						return selectorIndex, errors.New("comment was not terminated")
					}

					selectorIndex = endIndex + 2

					// Remove leading whitespace
					if len(tokens) == 0 {
						stripWhitespace(0)
					}

					break
				}

				var namespace *string
				var name string
				var err error

				if firstChar == Asterisk {
					selectorIndex += 1
					name = "*"
				} else if firstChar == Pipe {
					name = ""

					if selector[selectorIndex+1] == Pipe {
						err = addTraversal(SelectorTypeColumnCombinator)
						if err != nil {
							return selectorIndex, err
						}
						stripWhitespace(2)
						break
					}
				} else if reName.MatchString(selector[selectorIndex:]) {
					name, err = getName(0)
					if err != nil {
						return selectorIndex, err
					}
				} else {
					break loop
				}

				if selectorIndex < len(selector) && selector[selectorIndex] == Pipe && selector[selectorIndex+1] != Pipe {
					namespace = String(name)
					if selector[selectorIndex+1] == Asterisk {
						name = "*"
						selectorIndex += 2
					} else {
						name, err = getName(1)
						if err != nil {
							return selectorIndex, err
						}
					}
				}

				if name == "*" {
					tokens = append(tokens, &Selector{Type: SelectorTypeUniversal, Namespace: namespace})
				} else {
					tokens = append(tokens, &Selector{Type: SelectorTypeTag, Name: name, Namespace: namespace})
				}
			}
		}
	}

	err := finalizeSubselector()
	return selectorIndex, err
}
