package parser

type SelectorType string

const (
	SelectorTypeAttribute     SelectorType = "attribute"
	SelectorTypePseudo        SelectorType = "pseudo"
	SelectorTypePseudoElement SelectorType = "pseudo-element"
	SelectorTypeTag           SelectorType = "tag"
	SelectorTypeUniversal     SelectorType = "universal"

	// Traversals
	SelectorTypeAdjacent         SelectorType = "adjacent"
	SelectorTypeChild            SelectorType = "child"
	SelectorTypeDescendant       SelectorType = "descendant"
	SelectorTypeParent           SelectorType = "parent"
	SelectorTypeSibling          SelectorType = "sibling"
	SelectorTypeColumnCombinator SelectorType = "column-combinator"
)

type IgnoreCaseMode int

const (
	IgnoreCaseModeUnknown IgnoreCaseMode = iota
	IgnoreCaseModeQuirksMode
	IgnoreCaseModeIgnoreCase
	IgnoreCaseModeCaseSensitive
)

type AttributeAction string

const (
	AttributeActionAny     AttributeAction = "any"
	AttributeActionElement AttributeAction = "element"
	AttributeActionEnd     AttributeAction = "end"
	AttributeActionEquals  AttributeAction = "equals"
	AttributeActionExists  AttributeAction = "exists"
	AttributeActionHyphen  AttributeAction = "hyphen"
	AttributeActionNot     AttributeAction = "not"
	AttributeActionStart   AttributeAction = "start"
)

type Selector struct {
	Type       SelectorType
	Name       string
	Namespace  *string
	Data       *string
	Action     AttributeAction
	IgnoreCase IgnoreCaseMode
	Children   [][]*Selector
}
