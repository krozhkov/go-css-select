package query

import (
	_ "embed"
	"fmt"
	"slices"
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/sizzle.html
var sizzle string

//go:embed testdata/fries.xml
var fries string

func TestSizzle(t *testing.T) {
	loadDoc := func() []*dom.Node {
		document, err := dom.ParseDocument(sizzle, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
		require.NoError(t, err)
		return document.Children
	}

	/*createWithFriesXML := func() []*dom.Node {
		document := dom.ParseDocument(fries, &parser.ParserOptions{XmlMode: true})
		return document.Children
	}*/

	test := func(t *testing.T, selector string, expectedIds []string, context []*dom.Node) {
		actual, err := SelectAll(selector, context, nil)
		require.NoError(t, err)

		actualIds := internal.MapFunc(actual, func(e *dom.Node) string {
			return domutils.GetAttributeValue(e, "id")
		})

		// Should not contain falsy values
		assert.Equal(t, actualIds, expectedIds)
	}

	t.Run("pseudo - has", func(t *testing.T) {
		document := loadDoc()

		// Basic test
		test(t, "p:has(a)", []string{"firstp", "ap", "en", "sap"}, document)
		// Basic test (irrelevant whitespace)
		test(t, "p:has( a )", []string{"firstp", "ap", "en", "sap"}, document)
		// Nested with overlapping candidates
		test(t, "#qunit-fixture div:has(div:has(div:not([id])))", []string{"moretests", "t2037"}, document)
	})

	t.Run("pseudo - misc", func(t *testing.T) {
		document := loadDoc()

		// Headers
		test(t, ":header", []string{"qunit-header", "qunit-banner", "qunit-userAgent"}, document)
		// Headers(case-insensitive)
		test(t, ":Header", []string{"qunit-header", "qunit-banner", "qunit-userAgent"}, document)
		// Multiple matches with the same context (cache check)
		test(t, "#form select:has(option:first-child:contains('o'))", []string{
			"select1",
			"select2",
			"select3",
			"select4",
		}, document)

		matches, err := SelectAll("#qunit-fixture :not(:has(:has(*)))", document, nil)
		require.NoError(t, err)
		// All not grandparents
		assert.Len(t, matches, 175)

		select1 := domutils.GetElementById("select1", document, true)
		// Has Option Matches
		result, err := Is(select1, ":has(option)", nil)
		require.NoError(t, err)
		assert.True(t, result)

		// Empty string contains
		matches, err = SelectAll("a:contains('')", document, nil)
		require.NoError(t, err)
		assert.Len(t, matches, 18)

		// Text Contains
		test(t, "a:contains(Google)", []string{"google", "groups"}, document)
		// Text Contains
		test(t, "a:contains(Google Groups)", []string{"groups"}, document)

		// Text Contains
		test(t, "a:contains('Google Groups (Link)')", []string{"groups"}, document)
		// Text Contains
		test(t, "a:contains(\"(Link)\")", []string{"groups"}, document)
		// Text Contains
		test(t, "a:contains(Google Groups (Link))", []string{"groups"}, document)
		// Text Contains
		test(t, "a:contains((Link))", []string{"groups"}, document)

		tmp := dom.NewElement("div", orderedmap.NewOrderedMapWithElements(&dom.Attribute{Key: "id", Value: "tmp_input"}), nil, "")
		body := domutils.GetElementsByTagName("body", document, true, 1)[0]
		domutils.AppendChild(body, tmp)

		for _, typ := range []string{"button", "submit", "reset"} {
			html := fmt.Sprintf("<input id='input_%s' type='%s'/><button id='button_%s' type='%s'>test</button>", typ, typ, typ, typ)
			els := slices.Clone(parseDOM(html, false)) // Create a copy of the array, so that `appendChild` doesn't remove the elements.

			for _, el := range els {
				domutils.AppendChild(tmp, el)
			}

			// Input Buttons :${type}
			test(t, "#tmp_input :"+typ, []string{"input_" + typ, "button_" + typ}, document)

			// Input Matches :${type}
			result, err := Is(els[0], ":"+typ, nil)
			require.NoError(t, err)
			assert.True(t, result)
			// Button Matches :${type}
			result, err = Is(els[1], ":"+typ, nil)
			require.NoError(t, err)
			assert.True(t, result)
		}

		domutils.RemoveElement(tmp)

		// Caching system tolerates recursive selection
		test(t,
			"[id='select1'] *:not(:last-child), [id='select2'] *:not(:last-child)",
			[]string{
				"option1a",
				"option1b",
				"option1c",
				"option2a",
				"option2b",
				"option2c",
			},
			[]*dom.Node{domutils.GetElementById("qunit-fixture", document, true)},
		)

		/*
		 * Tokenization edge cases
		 */
		// Sequential pseudos
		test(t, "#qunit-fixture p:has(:contains(mark)):has(code)", []string{"ap"}, document)
		test(t,
			"#qunit-fixture p:has(:contains(mark)):has(code):contains(This link)",
			[]string{"ap"}, document,
		)

		// Pseudo argument containing ')'
		test(t, "p:has(>a.GROUPS[src!=')'])", []string{"ap"}, document)
		test(t, "p:has(>a.GROUPS[src!=')'])", []string{"ap"}, document)
		// Pseudo followed by token containing ')'
		test(t, "p:contains(id=\"foo\")[id!=\\)]", []string{"sndp"}, document)
		test(t, "p:contains(id=\"foo\")[id!=')']", []string{"sndp"}, document)

		// Multi-pseudo
		test(t, "#ap:has(*), #ap:has(*)", []string{"ap"}, document)
		// Multi-pseudo with leading nonexistent id
		test(t, "#nonexistent:has(*), #ap:has(*)", []string{"ap"}, document)

		// Tokenization stressor
		test(t,
			"a[class*=blog]:not(:has(*, :contains(!)), :contains(!)), br:contains(]), p:contains(]), :not(:empty):not(:parent)",
			[]string{"ap", "mark", "yahoo", "simon"}, document,
		)
	})

	t.Run("pseudo - :not", func(t *testing.T) {
		document := loadDoc()

		// Not
		test(t, "a.blog:not(.link)", []string{"mark"}, document)

		// Not - multiple
		test(t, "#form option:not(:contains(Nothing),#option1b,:selected)", []string{
			"option1c",
			"option1d",
			"option2b",
			"option2c",
			"option3d",
			"option3e",
			"option4e",
			"option5b",
			"option5c",
		}, document)
		// Not - recursive
		test(t, "#form option:not(:not(:selected))[id^='option3']", []string{
			"option3b",
			"option3c",
		}, document)

		// :not() failing interior
		test(t, "#qunit-fixture p:not(.foo)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not() failing interior
		test(t, "#qunit-fixture p:not(div.foo)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not() failing interior
		test(t, "#qunit-fixture p:not(p.foo)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not() failing interior
		test(t, "#qunit-fixture p:not(#blargh)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not() failing interior
		test(t, "#qunit-fixture p:not(div#blargh)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not() failing interior
		test(t, "#qunit-fixture p:not(p#blargh)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)

		// :not Multiple
		test(t, "#qunit-fixture p:not(a)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not Multiple
		test(t, "#qunit-fixture p:not( a )", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not Multiple
		test(t, "#qunit-fixture p:not( p )", []string{}, document)
		// :not Multiple
		test(t, "#qunit-fixture p:not(a, b)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not Multiple
		test(t, "#qunit-fixture p:not(a, b, div)", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// :not Multiple
		test(t, "p:not(p)", []string{}, document)
		// :not Multiple
		test(t, "p:not(a,p)", []string{}, document)
		// :not Multiple
		test(t, "p:not(p,a)", []string{}, document)
		// :not Multiple
		test(t, "p:not(a,p,b)", []string{}, document)
		// :not Multiple
		test(t, ":input:not(:image,:input,:submit)", []string{}, document)
		// :not Multiple
		test(t, "#qunit-fixture p:not(:has(a), :nth-child(1))", []string{"first"}, document)

		// No element not selector
		test(t, ".container div:not(.excluded) div", []string{}, document)

		// :not() Existing attribute
		test(t, "#form select:not([multiple])", []string{"select1", "select2", "select5"}, document)
		// :not() Equals attribute
		test(t, "#form select:not([name=select1])", []string{
			"select2",
			"select3",
			"select4",
			"select5",
		}, document)
		// :not() Equals quoted attribute
		test(t, "#form select:not([name='select1'])", []string{
			"select2",
			"select3",
			"select4",
			"select5",
		}, document)

		// :not() Multiple Class
		test(t, "#foo a:not(.blog)", []string{"yahoo", "anchor2"}, document)
		// :not() Multiple Class
		test(t, "#foo a:not(.link)", []string{"yahoo", "anchor2"}, document)
		// :not() Multiple Class
		test(t, "#foo a:not(.blog.link)", []string{"yahoo", "anchor2"}, document)

		// :not chaining (compound)
		test(t, "#qunit-fixture div[id]:not(:has(div, span)):not(:has(*))", []string{
			"nothiddendivchild",
			"divWithNoTabIndex",
		}, document)
		// :not chaining (with attribute)
		test(t, "#qunit-fixture form[id]:not([action$='formaction']):not(:button)", []string{
			"lengthtest",
			"name-tests",
			"testForm",
		}, document)
		// :not chaining (colon in attribute)
		test(t, "#qunit-fixture form[id]:not([action='form:action']):not(:button)", []string{
			"form",
			"lengthtest",
			"name-tests",
			"testForm",
		}, document)
		// :not chaining (colon in attribute and nested chaining)
		test(t,
			"#qunit-fixture form[id]:not([action='form:action']:button):not(:input)",
			[]string{"form", "lengthtest", "name-tests", "testForm"}, document,
		)
		// :not chaining
		test(t,
			"#form select:not(.select1):contains(Nothing) > option:not(option)",
			[]string{}, document,
		)
	})
}
