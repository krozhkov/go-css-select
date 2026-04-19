package parser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type selectorTest struct {
	selector string
	expected [][]*Selector
	message  string
}

var tests = []*selectorTest{
	// Tag names
	{
		selector: "div",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "simple tag",
	},
	{
		selector: "*",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeUniversal,
				},
			},
		},
		message: "universal",
	},

	// Traversal
	{
		selector: "div div",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
				{
					Type: SelectorTypeDescendant,
				},
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "descendant",
	},
	{
		selector: "div\t \n \tdiv",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
				{
					Type: SelectorTypeDescendant,
				},
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "descendant /w whitespace",
	},
	{
		selector: "div + div",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
				{
					Type: SelectorTypeAdjacent,
				},
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "adjacent",
	},
	{
		selector: "div ~ div",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
				{
					Type: SelectorTypeSibling,
				},
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "adjacent",
	},
	{
		selector: "p < div",
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "p",
				},
				{
					Type: SelectorTypeParent,
				},
				{
					Type: SelectorTypeTag,
					Name: "div",
				},
			},
		},
		message: "parent",
	},

	// Escaped whitespace
	{
		selector: `#\  > a `,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "id",
					IgnoreCase: IgnoreCaseModeQuirksMode,
					Data:       String(" "),
				},
				{
					Type: SelectorTypeChild,
				},
				{
					Type: SelectorTypeTag,
					Name: "a",
				},
			},
		},
		message: "Space between escaped space and combinator",
	},
	{
		selector: `.\  `,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "class",
					Action:     AttributeActionElement,
					IgnoreCase: IgnoreCaseModeQuirksMode,
					Data:       String(" "),
				},
			},
		},
		message: "Space after escaped space",
	},
	{
		selector: `.m™²³`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "class",
					Action:     AttributeActionElement,
					IgnoreCase: IgnoreCaseModeQuirksMode,
					Data:       String("m™²³"),
				},
			},
		},
		message: "Special charecters in selector",
	},
	{
		selector: `\61 `,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "a",
				},
			},
		},
		message: "Numeric escape with space (BMP)",
	},
	{
		selector: `\1d306\01d306`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "\U0001D306\U0001D306",
				},
			},
		},
		message: "Numeric escape (outside BMP)",
	},
	{
		selector: `#\26 B`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "id",
					IgnoreCase: IgnoreCaseModeQuirksMode,
					Data:       String("&B"),
				},
			},
		},
		message: "id selector with escape sequence",
	},

	// Attributes
	{
		selector: `[name^="foo["]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionStart,
					Data:       String("foo["),
				},
			},
		},
		message: "quoted attribute",
	},
	{
		selector: `[name^="foo[bar]"]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionStart,
					Data:       String("foo[bar]"),
				},
			},
		},
		message: "quoted attribute",
	},
	{
		selector: `[name$="[bar]"]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionEnd,
					Data:       String("[bar]"),
				},
			},
		},
		message: "quoted attribute",
	},
	{
		selector: `[href *= "google"]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "href",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionAny,
					Data:       String("google"),
				},
			},
		},
		message: "quoted attribute with spaces",
	},
	{
		selector: "[value=\"\nsome text\n\"]",
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "value",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionEquals,
					Data:       String("\nsome text\n"),
				},
			},
		},
		message: "quoted attribute with internal newline",
	},
	{
		selector: `[name=foo\.baz]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionEquals,
					Data:       String("foo.baz"),
				},
			},
		},
		message: "attribute with escaped dot",
	},
	{
		selector: `[name=foo\[bar\]]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Action:     AttributeActionEquals,
					Data:       String("foo[bar]"),
				},
			},
		},
		message: "attribute with escaped square brackets",
	},
	{
		selector: `[xml\:test]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "xml:test",
					Action:     AttributeActionExists,
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String(""),
				},
			},
		},
		message: "escaped attribute",
	},
	{
		selector: `[name='foo ~ < > , bar' i]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Name:       "name",
					Action:     AttributeActionEquals,
					IgnoreCase: IgnoreCaseModeIgnoreCase,
					Data:       String("foo ~ < > , bar"),
				},
			},
		},
		message: "attribute with previously normalized characters",
	},

	// ID starting with a dot
	{
		selector: `#.identifier`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "id",
					IgnoreCase: IgnoreCaseModeQuirksMode,
					Data:       String(".identifier"),
				},
			},
		},
		message: "ID starting with a dot",
	},

	// Pseudo elements
	{
		selector: `::foo`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudoElement,
					Name: "foo",
				},
			},
		},
		message: "pseudo-element",
	},
	{
		selector: `::foo()`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudoElement,
					Name: "foo",
					Data: String(""),
				},
			},
		},
		message: "pseudo-element",
	},
	{
		selector: `::foo(bar())`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudoElement,
					Name: "foo",
					Data: String("bar()"),
				},
			},
		},
		message: "pseudo-element",
	},

	// Pseudo selectors
	{
		selector: `:foo`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "foo",
				},
			},
		},
		message: "pseudo selector without any data",
	},
	{
		selector: `:bar(baz)`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "bar",
					Data: String("baz"),
				},
			},
		},
		message: "pseudo selector with data",
	},
	{
		selector: `:contains("(foo)")`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "contains",
					Data: String("(foo)"),
				},
			},
		},
		message: "pseudo selector with data",
	},
	{
		selector: `:where(a)`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "where",
					Children: [][]*Selector{
						{
							{
								Type: SelectorTypeTag,
								Name: "a",
							},
						},
					},
				},
			},
		},
		message: "pseudo selector with data",
	},
	{
		selector: `:contains("(a((foo\\\))))")`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "contains",
					Data: String("(a((foo))))"),
				},
			},
		},
		message: "pseudo selector with escaped data",
	},
	{
		selector: `:icontains('')`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "icontains",
					Data: String(""),
				},
			},
		},
		message: "pseudo selector with quote-stripped data",
	},

	// Multiple selectors
	{
		selector: `a , b`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "a",
				},
			},
			{
				{
					Type: SelectorTypeTag,
					Name: "b",
				},
			},
		},
		message: "multiple selectors",
	},
	{
		selector: `:host(h1, p)`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypePseudo,
					Name: "host",
					Children: [][]*Selector{
						{
							{
								Type: SelectorTypeTag,
								Name: "h1",
							},
						},
						{
							{
								Type: SelectorTypeTag,
								Name: "p",
							},
						},
					},
				},
			},
		},
		message: "pseudo selector with data",
	},

	/*
	 * Bad attributes (taken from Sizzle)
	 * https://github.com/jquery/sizzle/blob/af163873d7cdfc57f18b16c04b1915209533f0b1/test/unit/selector.js#L602-L651
	 */
	{
		selector: `[id=types_all]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "id",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("types_all"),
				},
			},
		},
		message: "Underscores don't need escaping",
	},
	{
		selector: `[name=foo\ bar]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("foo bar"),
				},
			},
		},
		message: "Escaped space",
	},
	{
		selector: `[name=foo\.baz]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("foo.baz"),
				},
			},
		},
		message: "Escaped dot",
	},
	{
		selector: `[name=foo\[baz\]]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "name",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("foo[baz]"),
				},
			},
		},
		message: "Escaped brackets",
	},
	{
		selector: `[data-attr='foo_baz\']']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("foo_baz']"),
				},
			},
		},
		message: "Escaped quote + right bracket",
	},
	{
		selector: `[data-attr='\'']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("'"),
				},
			},
		},
		message: "Quoted quote",
	},
	{
		selector: `[data-attr='\\']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\"),
				},
			},
		},
		message: "Quoted backslash",
	},
	{
		selector: `[data-attr='\\\'']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\'"),
				},
			},
		},
		message: "Quoted backslash quote",
	},
	{
		selector: `[data-attr='\\\\']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\\\"),
				},
			},
		},
		message: "Quoted backslash backslash",
	},
	{
		selector: `[data-attr='\5C\\']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\\\"),
				},
			},
		},
		message: "Quoted backslash backslash (numeric escape)",
	},
	{
		selector: `[data-attr='\5C \\']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\\\"),
				},
			},
		},
		message: "Quoted backslash backslash (numeric escape with trailing space)",
	},
	{
		selector: "[data-attr='\\5C\t\\\\']",
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\\\\"),
				},
			},
		},
		message: "Quoted backslash backslash (numeric escape with trailing tab)",
	},
	{
		selector: `[data-attr='\04e00']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\u4E00"),
				},
			},
		},
		message: "Long numeric escape (BMP)",
	},
	{
		selector: `[data-attr='\01D306A']`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "data-attr",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String("\U0001D306A"),
				},
			},
		},
		message: "Long numeric escape (non-BMP)",
	},
	{
		selector: `fOo[baR]`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "fOo",
				},
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionExists,
					Name:       "baR",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String(""),
				},
			},
		},
		message: "Mixed case tag and attribute name",
	},

	// Namespaces
	{
		selector: `foo|bar`,
		expected: [][]*Selector{
			{
				{
					Type:      SelectorTypeTag,
					Namespace: String("foo"),
					Name:      "bar",
				},
			},
		},
		message: "basic tag namespace",
	},
	{
		selector: `*|bar`,
		expected: [][]*Selector{
			{
				{
					Type:      SelectorTypeTag,
					Namespace: String("*"),
					Name:      "bar",
				},
			},
		},
		message: "star tag namespace",
	},
	{
		selector: `|bar`,
		expected: [][]*Selector{
			{
				{
					Type:      SelectorTypeTag,
					Name:      "bar",
					Namespace: String(""),
				},
			},
		},
		message: "without namespace",
	},
	{
		selector: `*|*`,
		expected: [][]*Selector{
			{
				{
					Type:      SelectorTypeUniversal,
					Namespace: String("*"),
				},
			},
		},
		message: "universal with namespace",
	},
	{
		selector: `[foo|bar]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionExists,
					Namespace:  String("foo"),
					Name:       "bar",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String(""),
				},
			},
		},
		message: "basic attribute namespace, existential",
	},
	{
		selector: `[|bar]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionExists,
					Name:       "bar",
					IgnoreCase: IgnoreCaseModeUnknown,
					Data:       String(""),
				},
			},
		},
		message: "without namespace, existential",
	},
	{
		selector: `[foo|bar='baz' i]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Namespace:  String("foo"),
					Name:       "bar",
					IgnoreCase: IgnoreCaseModeIgnoreCase,
					Data:       String("baz"),
				},
			},
		},
		message: "basic attribute namespace, equality",
	},
	{
		selector: `[*|bar='baz' i]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Namespace:  String("*"),
					Name:       "bar",
					IgnoreCase: IgnoreCaseModeIgnoreCase,
					Data:       String("baz"),
				},
			},
		},
		message: "star attribute namespace",
	},
	{
		selector: `[type='a' S]`,
		expected: [][]*Selector{
			{
				{
					Type:       SelectorTypeAttribute,
					Action:     AttributeActionEquals,
					Name:       "type",
					IgnoreCase: IgnoreCaseModeCaseSensitive,
					Data:       String("a"),
				},
			},
		},
		message: "case-sensitive attribute selector",
	},
	{
		selector: `foo || bar`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "foo",
				},
				{
					Type: SelectorTypeColumnCombinator,
				},
				{
					Type: SelectorTypeTag,
					Name: "bar",
				},
			},
		},
		message: "column combinator",
	},
	{
		selector: `foo||bar`,
		expected: [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "foo",
				},
				{
					Type: SelectorTypeColumnCombinator,
				},
				{
					Type: SelectorTypeTag,
					Name: "bar",
				},
			},
		},
		message: "column combinator without whitespace",
	},
}

var broken = []string{
	"[",
	"(",
	"{",
	"()",
	"<>",
	"{}",
	",",
	",a",
	"a,",
	"[id=012345678901234567890123456789",
	"input[name=foo b]",
	"input[name!foo]",
	"input[name|]",
	"input[name=']",
	"input[name=foo[baz]]",
	":has(\"p\")",
	":has(p",
	":foo(p()",
	"#",
	"##foo",
	"/*",
}

func TestOwnTests(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			parsed, err := Parse(tt.selector)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func TestBrokenSelectors(t *testing.T) {
	for _, selector := range broken {
		t.Run(fmt.Sprintf("should not parse — %s", selector), func(t *testing.T) {
			_, err := Parse(selector)
			assert.NotNil(t, err)
		})
	}
}

//go:embed __fixtures__/out.json
var testData []byte

func TestCollectedSelectors(t *testing.T) {
	var out map[string][][]*Selector

	err := json.Unmarshal(testData, &out)
	assert.Nil(t, err)

	for selector, expected := range out {
		t.Run(selector, func(t *testing.T) {
			parsed, err := Parse(selector)
			assert.Nil(t, err)
			assert.Equal(t, expected, parsed)
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("should ignore comments", func(t *testing.T) {
		parsed, err := Parse("/* comment1 */ /**/ foo /*comment2*/")
		assert.Nil(t, err)
		assert.Equal(t, [][]*Selector{
			{
				{
					Type: SelectorTypeTag,
					Name: "foo",
				},
			},
		}, parsed)
	})

	t.Run("should support legacy pseudo-elements with single colon", func(t *testing.T) {
		parsed, err := Parse(":before")
		assert.Nil(t, err)
		assert.Equal(t, [][]*Selector{
			{
				{
					Type: SelectorTypePseudoElement,
					Name: "before",
				},
			},
		}, parsed)
	})
}
