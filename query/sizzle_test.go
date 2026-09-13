package query

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-css-select/query/types"
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

	createWithFriesXML := func() []*dom.Node {
		document, err := dom.ParseDocument(fries, &parser.ParserOptions{XmlMode: true})
		require.NoError(t, err)
		return document.Children
	}

	test := func(t *testing.T, selector string, expectedIds []string, context []*dom.Node) {
		actual, err := SelectAll(selector, context, nil)
		require.NoError(t, err)

		actualIds := internal.MapFunc(actual, func(e *dom.Node) string {
			return domutils.GetAttributeValue(e, "id")
		})

		// Should not contain falsy values
		assert.Equal(t, actualIds, expectedIds)
	}

	testR := func(t *testing.T, selector string, expectedIds []string, context []*dom.Node) {
		actual, err := SelectAll(selector, context, &types.Options{RelativeSelector: types.OptYes})
		require.NoError(t, err)

		actualIds := internal.MapFunc(actual, func(e *dom.Node) string {
			return domutils.GetAttributeValue(e, "id")
		})

		// Should not contain falsy values
		assert.Equal(t, actualIds, expectedIds)
	}

	t.Run("element", func(t *testing.T) {
		document := loadDoc()

		// Empty selector returns an empty array
		nodes, err := SelectAll("", document, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 0)
		// Text element as context fails silently
		nodes, err = SelectAll("div", []*dom.Node{dom.NewText("")}, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 0)
		form := domutils.GetElementById("form", document, true)
		// Empty string passed to matchesSelector does not match
		result, err := Is(form, "", nil)
		require.NoError(t, err)
		assert.False(t, result)
		// Empty selector returns an empty array
		nodes, err = SelectAll(" ", document, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 0)
		// Empty selector returns an empty array
		nodes, err = SelectAll("\t", document, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 0)

		// Select all
		nodes, err = SelectAll("*", document, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(nodes), 30)
		all, err := SelectAll("*", document, nil)
		good := internal.Every(all, func(el *dom.Node) bool { return el.NodeType != 8 })
		// Select all elements, no comment nodes
		assert.True(t, good)
		// Element Selector
		test(t, "html", []string{"html"}, document)
		// Element Selector
		test(t, "body", []string{"body"}, document)
		// Element Selector
		test(t, "#qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)

		// Leading space
		test(t, " #qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Leading tab
		test(t, "\t#qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Leading carriage return
		test(t, "\r#qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Leading line feed
		test(t, "\n#qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Leading form feed
		test(t, "\f#qunit-fixture p", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Trailing space
		test(t, "#qunit-fixture p ", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Trailing tab
		test(t, "#qunit-fixture p\t", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Trailing carriage return
		test(t, "#qunit-fixture p\r", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Trailing line feed
		test(t, "#qunit-fixture p\n", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)
		// Trailing form feed
		test(t, "#qunit-fixture p\f", []string{"firstp", "ap", "sndp", "en", "sap", "first"}, document)

		// Parent Element
		test(t, "dl ol", []string{"empty", "listWithTabIndex"}, document)
		// Parent Element (non-space descendant combinator)
		test(t, "dl\tol", []string{"empty", "listWithTabIndex"}, document)
		obj1 := domutils.GetElementById("object1", document, true)
		// Object/param as context
		nodes, err = SelectAll("param", []*dom.Node{obj1}, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 2)

		// Finding selects with a context.
		test(t,
			"select",
			[]string{"select1", "select2", "select3", "select4", "select5"},
			[]*dom.Node{form},
		)

		/*
		 * Check for unique-ness and sort order
		 * Check for duplicates: p, div p
		 */
		one, err := SelectAll("p, div p", document, nil)
		require.NoError(t, err)
		another, err := SelectAll("p", document, nil)
		require.NoError(t, err)
		assert.Equal(t, one, another)

		// Checking sort order
		test(t, "h2, h1", []string{"qunit-header", "qunit-banner", "qunit-userAgent"}, document)
		// Checking sort order
		test(t, "#qunit-fixture p, #qunit-fixture p a", []string{
			"firstp",
			"simon1",
			"ap",
			"google",
			"groups",
			"anchor1",
			"mark",
			"sndp",
			"en",
			"yahoo",
			"sap",
			"anchor2",
			"simon",
			"first",
		}, document)

		// Test Conflict ID
		lengthtest := []*dom.Node{domutils.GetElementById("lengthtest", document, true)}
		// Finding element with id of ID.
		test(t, "#idTest", []string{"idTest"}, lengthtest)
		// Finding element with id of ID.
		test(t, "[name='id']", []string{"idTest"}, lengthtest)
		// Finding elements with id of ID.
		test(t, "input[id='idTest']", []string{"idTest"}, lengthtest)

		siblingTest := domutils.GetElementById("siblingTest", document, true).Children
		// Element-rooted QSA does not select based on document context
		testR(t, "div em", []string{}, siblingTest)
		// Element-rooted QSA does not select based on document context
		testR(t, "div em, div em, div em:not(div em)", []string{}, siblingTest)
		// Escaped commas do not get treated with an id in element-rooted QSA
		testR(t, "div em, em\\,", []string{}, siblingTest)

		iframe := domutils.GetElementById("iframe", document, true)
		iframe.Children = parseDOM("<body><p id='foo'>bar</p></body>", false)
		for _, e := range iframe.Children {
			e.Parent = iframe
		}
		// Other document as context
		nodes, err = SelectAll("p:contains(bar)", iframe.Children, nil)
		require.NoError(t, err)
		assert.Equal(t, nodes, []*dom.Node{domutils.GetElementById("foo", iframe.Children, true)})
		iframe.Children = []*dom.Node{}

		markup := ""
		for i := 0; i < 100; i++ {
			markup = "<div>" + markup + "</div>"
		}
		html := parseDOM(markup, false)[0]
		body := domutils.GetElementsByTagName("body", document, true, 1)[0]
		domutils.AppendChild(body, html)
		// No stack or performance problems with large amounts of descendents
		nodes, err = SelectAll("body div div div", document, nil)
		require.NoError(t, err)
		assert.Greater(t, len(nodes), 0)
		domutils.RemoveElement(html)

		// Real use case would be using .watch in browsers with window.watch (see Issue #157)
		elem := dom.NewElement(strings.ToLower("toString"), orderedmap.NewOrderedMap[string, string](), nil, "")
		elem.Attribs.Set("id", "toString")
		fixture := domutils.GetElementById("qunit-fixture", document, true)
		domutils.AppendChild(fixture, elem)
		// Element name matches Object.prototype property
		test(t, "tostring#toString", []string{"toString"}, document)
		domutils.RemoveElement(elem)
	})

	t.Run("XML Document Selectors", func(t *testing.T) {
		xml := createWithFriesXML()

		// Element Selector with underscore
		nodes, err := SelectAll("foo_bar", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Class selector
		nodes, err = SelectAll(".component", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector for class
		nodes, err = SelectAll("[class*=component]", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector with name
		nodes, err = SelectAll("property[name=prop2]", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector with name
		nodes, err = SelectAll("[name=prop2]", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector with ID
		nodes, err = SelectAll("#seite1", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector with ID
		nodes, err = SelectAll("component#seite1", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		// Attribute selector filter with ID
		nodes, err = SelectAll("component", xml, nil)
		require.NoError(t, err)
		filtered := internal.FilterFunc(nodes, func(n *dom.Node) bool {
			result, err := Is(n, "#seite1", nil)
			require.NoError(t, err)
			return result
		})
		assert.Len(t, filtered, 1)
		// Descendent selector and dir caching
		nodes, err = SelectAll("meta property thing", xml, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 2)
		// Check for namespaced element
		filtered = internal.FilterFunc(xml, func(n *dom.Node) bool { return n.Type == "tag" })
		last := filtered[len(filtered)-1]
		result, err := Is(last, "soap\\:Envelope", &types.Options{XmlMode: types.OptYes})
		require.NoError(t, err)
		assert.True(t, result)

		doc, err := dom.ParseDocument("<?xml version='1.0' encoding='UTF-8'?><root><elem id='1'/></root>", &parser.ParserOptions{XmlMode: true})
		require.NoError(t, err)
		// Non-qSA path correctly handles numeric ids (jQuery #14142)
		nodes, err = SelectAll("elem:not(:has(*))", doc.Children, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
	})

	t.Run("broken", func(t *testing.T) {
		document := loadDoc()

		broken := func(t *testing.T, selector string) {
			_, err := Compile(selector, nil, nil)
			assert.Error(t, err)
		}

		broken(t, "[")
		broken(t, "(")
		broken(t, "{")
		broken(t, "()")
		broken(t, "<>")
		broken(t, "{}")
		broken(t, ",")
		broken(t, ",a")
		broken(t, "a,")
		// Hangs on IE 9 if regular expression is inefficient
		broken(t, "[id=012345678901234567890123456789")
		// Doesn't exist
		broken(t, ":visble")
		broken(t, ":nth-child")
		/*
		 * Sigh again. IE 9 thinks this is also a real selector
		 * Not super critical that we fix this case
		 */
		broken(t, ":nth-child(-)")
		/*
		 * Sigh. WebKit thinks this is a real selector in qSA
		 * They've already fixed this and it'll be coming into
		 * Current browsers soon. Currently, Safari 5.0 still has this problem
		 */
		broken(t, ":nth-child(asdf)")
		broken(t, ":nth-child(2n+-0)")
		broken(t, ":nth-child(2+0)")
		broken(t, ":nth-child(- 1n)")
		broken(t, ":nth-child(-1 n)")
		broken(t, ":first-child(n)")
		broken(t, ":last-child(n)")
		broken(t, ":only-child(n)")
		broken(t, ":nth-last-last-child(1)")
		broken(t, ":first-last-child")
		broken(t, ":last-last-child")
		broken(t, ":only-last-child")

		// Make sure attribute value quoting works correctly. See: #6093
		for _, node := range parseDOM(
			"<input type='hidden' value='2' name='foo.baz' id='attrbad1'/><input type='hidden' value='2' name='foo[baz]' id='attrbad2'/>",
			false,
		) {
			domutils.AppendChild(domutils.GetElementById("form", document, true), node)
		}

		// Shouldn't be matching those inner brackets
		broken(t, "input[name=foo[baz]]")
	})

	t.Run("id", func(t *testing.T) {
		document := loadDoc()

		// ID Selector
		test(t, "#body", []string{"body"}, document)
		// ID Selector w/ Element
		test(t, "body#body", []string{"body"}, document)
		// ID Selector w/ Element
		test(t, "ul#first", []string{}, document)
		// ID selector with existing ID descendant
		test(t, "#firstp #simon1", []string{"simon1"}, document)
		// ID selector with non-existant descendant
		test(t, "#firstp #foobar", []string{}, document)
		// ID selector using UTF8
		test(t, "#台北Táiběi", []string{"台北Táiběi"}, document)
		// Multiple ID selectors using UTF8
		test(t, "#台北Táiběi, #台北", []string{"台北Táiběi", "台北"}, document)
		// Descendant ID selector using UTF8
		test(t, "div #台北", []string{"台北"}, document)
		// Child ID selector using UTF8
		test(t, "form > #台北", []string{"台北"}, document)

		// Escaped ID
		test(t, "#foo\\:bar", []string{"foo:bar"}, document)
		// Escaped ID with descendent
		test(t, "#foo\\:bar span:not(:input)", []string{"foo_descendent"}, document)
		// Escaped ID
		test(t, "#test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)
		// Descendant escaped ID
		test(t, "div #foo\\:bar", []string{"foo:bar"}, document)
		// Descendant escaped ID
		test(t, "div #test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)
		// Child escaped ID
		test(t, "form > #foo\\:bar", []string{"foo:bar"}, document)
		// Child escaped ID
		test(t, "form > #test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)

		fiddle := parseDOM(
			"<div id='fiddle\\Foo'><span id='fiddleSpan'></span></div>",
			false,
		)[0]
		domutils.AppendChild(domutils.GetElementById("qunit-fixture", document, true), fiddle)
		// Escaped ID as context
		node, err := SelectOne("#fiddle\\\\Foo", document, nil)
		require.NoError(t, err)
		test(t,
			"> span",
			[]string{"fiddleSpan"},
			[]*dom.Node{node},
		)

		domutils.RemoveElement(fiddle)

		// ID Selector, child ID present
		test(t, "#form > #radio1", []string{"radio1"}, document) // Bug #267
		// ID Selector, not an ancestor ID
		test(t, "#form #first", []string{}, document)
		// ID Selector, not a child ID
		test(t, "#form > #option1a", []string{}, document)

		// All Children of ID
		test(t, "#foo > *", []string{"sndp", "en", "sap"}, document)
		// All Children of ID with no children
		test(t, "#firstUL > *", []string{}, document)

		// ID selector with same value for a name attribute
		node, err = SelectOne("#tName1", document, nil)
		require.NoError(t, err)
		assert.Equal(t, node.Attribs.GetOrDefault("id", ""), "tName1")
		// ID selector non-existing but name attribute on an A tag
		test(t, "#tName2", []string{}, document)
		// Leading ID selector non-existing but name attribute on an A tag
		test(t, "#tName2 span", []string{}, document)
		// Leading ID selector existing, retrieving the child
		test(t, "#tName1 span", []string{"tName1-span"}, document)
		// Ending with ID
		node, err = SelectOne("div > div #tName1", document, nil)
		require.NoError(t, err)
		another, err := SelectOne("#tName1-span", document, nil)
		require.NoError(t, err)
		assert.Equal(t, node.Attribs.GetOrDefault("id", ""), another.Parent.Attribs.GetOrDefault("id", ""))

		for _, node := range parseDOM("<a id='backslash\\foo'></a>", false) {
			domutils.AppendChild(domutils.GetElementById("form", document, true), node)
		}

		// ID Selector contains backslash
		test(t, "#backslash\\\\foo", []string{"backslash\\foo"}, document)

		// ID Selector on Form with an input that has a name of 'id'
		test(t, "#lengthtest", []string{"lengthtest"}, document)

		// ID selector with non-existant ancestor
		test(t, "#asdfasdf #foobar", []string{}, document) // Bug #986

		// ID selector within the context of another element
		body := domutils.GetElementsByTagName("body", document, true, 1)[0]
		test(t, "div#form", []string{}, []*dom.Node{body})

		// Underscore ID
		test(t, "#types_all", []string{"types_all"}, document)
		// Dash ID
		test(t, "#qunit-fixture", []string{"qunit-fixture"}, document)

		// ID with weird characters in it
		test(t, "#name\\+value", []string{"name+value"}, document)
	})

	t.Run("class", func(t *testing.T) {
		document := loadDoc()

		// Class Selector
		test(t, ".blog", []string{"mark", "simon"}, document)
		// Class Selector
		test(t, ".GROUPS", []string{"groups"}, document)
		// Class Selector
		test(t, ".blog.link", []string{"simon"}, document)
		// Class Selector w/ Element
		test(t, "a.blog", []string{"mark", "simon"}, document)
		// Parent Class Selector
		test(t, "p .blog", []string{"mark", "simon"}, document)

		// Class selector using UTF8
		test(t, ".台北Táiběi", []string{"utf8class1"}, document)
		// Class selector using UTF8
		test(t, ".台北", []string{"utf8class1", "utf8class2"}, document)
		// Class selector using UTF8
		test(t, ".台北Táiběi.台北", []string{"utf8class1"}, document)
		// Class selector using UTF8
		test(t, ".台北Táiběi, .台北", []string{"utf8class1", "utf8class2"}, document)
		// Descendant class selector using UTF8
		test(t, "div .台北Táiběi", []string{"utf8class1"}, document)
		// Child class selector using UTF8
		test(t, "form > .台北Táiběi", []string{"utf8class1"}, document)

		// Escaped Class
		test(t, ".foo\\:bar", []string{"foo:bar"}, document)
		// Escaped Class
		test(t, ".test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)
		// Descendant escaped Class
		test(t, "div .foo\\:bar", []string{"foo:bar"}, document)
		// Descendant escaped Class
		test(t, "div .test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)
		// Child escaped Class
		test(t, "form > .foo\\:bar", []string{"foo:bar"}, document)
		// Child escaped Class
		test(t, "form > .test\\.foo\\[5\\]bar", []string{"test.foo[5]bar"}, document)

		div := dom.NewElement("div", orderedmap.NewOrderedMap[string, string](), nil, "")
		div.Children = parseDOM("<div class='test e'></div><div class='test'></div>", false)
		for _, e := range div.Children {
			e.Parent = div
		}

		// Finding a second class.
		nodes, err := SelectAll(".e", []*dom.Node{div}, nil)
		require.NoError(t, err)
		assert.Equal(t, nodes, []*dom.Node{div.Children[0]})

		lastChild := div.Children[len(div.Children)-1]
		lastChild.Attribs.Set("class", "e")

		// Finding a modified class.
		nodes, err = SelectAll(".e", []*dom.Node{div}, nil)
		require.NoError(t, err)
		assert.Equal(t, nodes, []*dom.Node{div.Children[0], lastChild})

		// .null does not match an element with no class
		result, err := Is(div, ".null", nil)
		require.NoError(t, err)
		assert.False(t, result)
		// .null does not match an element with no class
		result, err = Is(div.Children[0], ".null div", nil)
		require.NoError(t, err)
		assert.False(t, result)
		div.Attribs.Set("class", "null")
		// .null matches element with class 'null'
		result, err = Is(div, ".null", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Caching system respects DOM changes
		result, err = Is(div.Children[0], ".null div", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Testing class on document doesn't error
		// result, err = Is(document, ".foo", nil)
		// require.NoError(t, err)
		// assert.False(t, result)
		lastChild.Attribs.Set("class", lastChild.Attribs.GetOrDefault("class", "")+" hasOwnProperty toString")
		// Testing class on global object doesn't error
		// expect(CSSselect.is(global, ".foo")).toBe(false);
		// Classes match Object.prototype properties
		nodes, err = SelectAll(".e.hasOwnProperty.toString", []*dom.Node{div}, nil)
		require.NoError(t, err)
		assert.Equal(t, nodes, []*dom.Node{lastChild})

		div2 := parseDOM(
			"<div><svg width='200' height='250' version='1.1' xmlns='http://www.w3.org/2000/svg'><rect x='10' y='10' width='30' height='30' class='foo'></rect></svg></div>",
			false,
		)
		// Class selector against SVG
		nodes, err = SelectAll(".foo", div2, nil)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
	})

	t.Run("name", func(t *testing.T) {
		document := loadDoc()

		// Name selector
		test(t, "input[name=action]", []string{"text1"}, document)
		// Name selector with single quotes
		test(t, "input[name='action']", []string{"text1"}, document)
		// Name selector with double quotes
		test(t, "input[name=\"action\"]", []string{"text1"}, document)

		// Name selector non-input
		test(t, "[name=example]", []string{"name-is-example"}, document)
		// Name selector non-input
		test(t, "[name=div]", []string{"name-is-div"}, document)
		// Name selector non-input
		test(t, "*[name=iframe]", []string{"iframe"}, document)

		// Name selector for grouped input
		test(t, "input[name='types[]']", []string{"types_all", "types_anime", "types_movie"}, document)

		form1 := domutils.GetElementById("form", document, true)
		// Name selector within the context of another element
		test(t, "input[name=action]", []string{"text1"}, []*dom.Node{form1})
		// Name selector for grouped form element within the context of another element
		test(t, "input[name='foo[bar]']", []string{"hidden2"}, []*dom.Node{form1})

		body := domutils.GetElementsByTagName("body", document, true, 1)[0]
		form2 := parseDOM("<form><input name='id'/></form>", false)[0]
		domutils.AppendChild(body, form2)

		// Make sure that rooted queries on forms (with possible expandos) work.
		result, err := SelectAll("input", []*dom.Node{form2}, nil)
		require.NoError(t, err)
		assert.Len(t, result, 1)

		domutils.RemoveElement(form2)

		// Find elements that have similar IDs
		test(t, "[name=tName1]", []string{"tName1ID"}, document)
		// Find elements that have similar IDs
		test(t, "[name=tName2]", []string{"tName2ID"}, document)
		// Find elements that have similar IDs
		test(t, "#tName2ID", []string{"tName2ID"}, document)
	})

	t.Run("multiple", func(t *testing.T) {
		document := loadDoc()

		// Comma Support
		test(t, "h2, #qunit-fixture p", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// Comma Support
		test(t, "h2 , #qunit-fixture p", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// Comma Support
		test(t, "h2 , #qunit-fixture p", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// Comma Support
		test(t, "h2,#qunit-fixture p", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// Comma Support
		test(t, "h2,#qunit-fixture p ", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
		// Comma Support
		test(t, "h2\t,\r#qunit-fixture p\n", []string{
			"qunit-banner",
			"qunit-userAgent",
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
	})

	t.Run("child and adjacent", func(t *testing.T) {
		document := loadDoc()

		// Child
		test(t, "p > a", []string{"simon1", "google", "groups", "mark", "yahoo", "simon"}, document)
		// Child
		test(t, "p> a", []string{"simon1", "google", "groups", "mark", "yahoo", "simon"}, document)
		// Child
		test(t, "p >a", []string{"simon1", "google", "groups", "mark", "yahoo", "simon"}, document)
		// Child
		test(t, "p>a", []string{"simon1", "google", "groups", "mark", "yahoo", "simon"}, document)
		// Child w/ Class
		test(t, "p > a.blog", []string{"mark", "simon"}, document)
		// All Children
		test(t, "code > *", []string{"anchor1", "anchor2"}, document)
		// All Grandchildren
		test(t, "p > * > *", []string{"anchor1", "anchor2"}, document)
		// Adjacent
		test(t, "#qunit-fixture a + a", []string{"groups", "tName2ID"}, document)
		// Adjacent
		test(t, "#qunit-fixture a +a", []string{"groups", "tName2ID"}, document)
		// Adjacent
		test(t, "#qunit-fixture a+ a", []string{"groups", "tName2ID"}, document)
		// Adjacent
		test(t, "#qunit-fixture a+a", []string{"groups", "tName2ID"}, document)
		// Adjacent
		test(t, "p + p", []string{"ap", "en", "sap"}, document)
		// Adjacent
		test(t, "p#firstp + p", []string{"ap"}, document)
		// Adjacent
		test(t, "p[lang=en] + p", []string{"sap"}, document)
		// Adjacent
		test(t, "a.GROUPS + code + a", []string{"mark"}, document)
		// Comma, Child, and Adjacent
		test(t, "#qunit-fixture a + a, code > a", []string{
			"groups",
			"anchor1",
			"anchor2",
			"tName2ID",
		}, document)
		// Element Preceded By
		test(t, "#qunit-fixture p ~ div", []string{
			"foo",
			"nothiddendiv",
			"moretests",
			"tabindex-tests",
			"liveHandlerOrder",
			"siblingTest",
		}, document)
		// Element Preceded By
		test(t, "#first ~ div", []string{
			"moretests",
			"tabindex-tests",
			"liveHandlerOrder",
			"siblingTest",
		}, document)
		// Element Preceded By
		test(t, "#groups ~ a", []string{"mark"}, document)
		// Element Preceded By
		test(t, "#length ~ input", []string{"idTest"}, document)
		// Element Preceded By
		test(t, "#siblingfirst ~ em", []string{"siblingnext", "siblingthird"}, document)
		// Element Preceded By (multiple)
		test(t, "#siblingTest em ~ em ~ em ~ span", []string{"siblingspan"}, document)
		// Element Preceded By, Containing
		test(t, "#liveHandlerOrder ~ div em:contains('1')", []string{"siblingfirst"}, document)

		siblingFirst := domutils.GetElementById("siblingfirst", document, true)

		// Element Preceded By with a context.
		test(t, "~ em", []string{"siblingnext", "siblingthird"}, []*dom.Node{siblingFirst})
		// Element Directly Preceded By with a context.
		test(t, "+ em", []string{"siblingnext"}, []*dom.Node{siblingFirst})

		en := domutils.GetElementById("en", document, true)
		// Compound selector with context, beginning with sibling test.
		test(t, "+ p, a", []string{"yahoo", "sap"}, []*dom.Node{en})
		// Compound selector with context, containing sibling test.
		test(t, "a, + p", []string{"yahoo", "sap"}, []*dom.Node{en})

		// Multiple combinators selects all levels
		test(t, "#siblingTest em *", []string{
			"siblingchild",
			"siblinggrandchild",
			"siblinggreatgrandchild",
		}, document)
		// Multiple combinators selects all levels
		test(t, "#siblingTest > em *", []string{
			"siblingchild",
			"siblinggrandchild",
			"siblinggreatgrandchild",
		}, document)
		// Multiple sibling combinators doesn't miss general siblings
		test(t, "#siblingTest > em:first-child + em ~ span", []string{"siblingspan"}, document)
		// Combinators are not skipped when mixing general and specific
		test(t, "#siblingTest > em:contains('x') + em ~ span", []string{}, document)

		// Parent div for next test is found via ID (#8310)
		result, err := SelectAll("#listWithTabIndex", document, nil)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		// Make sure the temporary id assigned by sizzle is cleared out (#8310)
		result, err = SelectAll("#__sizzle__", document, nil)
		require.NoError(t, err)
		assert.Len(t, result, 0)
		// Parent div for previous test is still found via ID (#8310)
		result, err = SelectAll("#listWithTabIndex", document, nil)
		require.NoError(t, err)
		assert.Len(t, result, 1)

		// Verify deep class selector
		test(t, "div.blah > p > a", []string{}, document)

		// No element deep selector
		test(t, "div.foo > span > a", []string{}, document)

		// Non-existant ancestors
		test(t, ".fototab > .thumbnails > a", []string{}, document)
		// Child of scope
		scope, err := SelectOne("#scopeTest", document, nil)
		require.NoError(t, err)
		test(t,
			":scope > label",
			[]string{"scopeTest--child"},
			[]*dom.Node{scope},
		)
	})

	t.Run("attributes", func(t *testing.T) {
		document := loadDoc()

		// Attribute Exists
		test(t, "#qunit-fixture a[title]", []string{"google"}, document)
		// Attribute Exists (case-insensitive)
		test(t, "#qunit-fixture a[TITLE]", []string{"google"}, document)
		// Attribute Exists
		test(t, "#qunit-fixture *[title]", []string{"google"}, document)
		// Attribute Exists
		test(t, "#qunit-fixture [title]", []string{"google"}, document)
		// Attribute Exists
		test(t, "#qunit-fixture a[ title ]", []string{"google"}, document)

		// Boolean attribute exists
		test(t, "#select2 option[selected]", []string{"option2d"}, document)
		// Boolean attribute equals
		test(t, "#select2 option[selected='selected']", []string{"option2d"}, document)

		// Attribute Equals
		test(t, "#qunit-fixture a[rel='bookmark']", []string{"simon1"}, document)
		// Attribute Equals
		test(t, "#qunit-fixture a[rel='bookmark']", []string{"simon1"}, document)
		// Attribute Equals
		test(t, "#qunit-fixture a[rel=bookmark]", []string{"simon1"}, document)
		// Attribute Equals
		test(t, "#qunit-fixture a[href='http://www.google.com/']", []string{"google"}, document)
		// Attribute Equals
		test(t, "#qunit-fixture a[ rel = 'bookmark' ]", []string{"simon1"}, document)
		// Attribute Equals Number
		test(t, "#qunit-fixture option[value=1]", []string{
			"option1b",
			"option2b",
			"option3b",
			"option4b",
			"option5c",
		}, document)
		// Attribute Equals Number
		test(t, "#qunit-fixture li[tabIndex=-1]", []string{"foodWithNegativeTabIndex"}, document)

		domutils.GetElementById("anchor2", document, true).Attribs.Set("href", "#2")
		// `href` Attribute
		test(t, "p a[href^=#]", []string{"anchor2"}, document)
		test(t, "p a[href*=#]", []string{"simon1", "anchor2"}, document)

		// `for` Attribute
		test(t, "form label[for]", []string{"label-for"}, document)
		// `for` Attribute in form
		test(t, "#form [for=action]", []string{"label-for"}, document)

		// Attribute containing []
		test(t, "input[name^='foo[']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name^='foo[bar]']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name*='[bar]']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name$='bar]']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name$='[bar]']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name$='foo[bar]']", []string{"hidden2"}, document)
		// Attribute containing []
		test(t, "input[name*='foo[bar]']", []string{"hidden2"}, document)

		// Without context, single-quoted attribute containing ','
		test(t, "input[data-comma='0,1']", []string{"el12087"}, document)
		// Without context, double-quoted attribute containing ','
		test(t, "input[data-comma=\"0,1\"]", []string{"el12087"}, document)
		// With context, single-quoted attribute containing ','
		test(t,
			"input[data-comma='0,1']",
			[]string{"el12087"},
			domutils.GetElementById("t12087", document, true).Children,
		)
		// With context, double-quoted attribute containing ','
		test(t,
			"input[data-comma=\"0,1\"]",
			[]string{"el12087"},
			domutils.GetElementById("t12087", document, true).Children,
		)

		// Multiple Attribute Equals
		test(t, "#form input[type='radio'], #form input[type='hidden']", []string{
			"radio1",
			"radio2",
			"hidden1",
		}, document)
		// Multiple Attribute Equals
		test(t, "#form input[type='radio'], #form input[type=\"hidden\"]", []string{
			"radio1",
			"radio2",
			"hidden1",
		}, document)
		// Multiple Attribute Equals
		test(t, "#form input[type='radio'], #form input[type=hidden]", []string{
			"radio1",
			"radio2",
			"hidden1",
		}, document)

		// Attribute selector using UTF8
		test(t, "span[lang=中文]", []string{"台北"}, document)

		// Attribute Begins With
		test(t, "a[href ^= 'http://www']", []string{"google", "yahoo"}, document)
		// Attribute Ends With
		test(t, "a[href $= 'org/']", []string{"mark"}, document)
		// Attribute Contains
		test(t, "a[href *= 'google']", []string{"google", "groups"}, document)
		// Attribute Is Not Equal
		test(t, "#ap a[hreflang!='en']", []string{"google", "groups", "anchor1"}, document)

		opt := domutils.GetElementById("option1a", document, true)
		opt.Attribs.Set("test", "")

		// Attribute Is Not Equal Matches
		result, err := Is(opt, "[id*=option1][type!=checkbox]", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Attribute With No Quotes Contains Matches
		result, err = Is(opt, "[id*=option1]", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Attribute With No Quotes No Content Matches
		result, err = Is(opt, "[test=]", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Attribute with empty string value does not match startsWith selector (^=)
		result, err = Is(opt, "[test^='']", nil)
		require.NoError(t, err)
		assert.False(t, result)
		// Attribute With No Quotes Equals Matches
		result, err = Is(opt, "[id=option1a]", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// Attribute With No Quotes Href Contains Matches
		result, err = Is(domutils.GetElementById("simon1", document, true), "a[href*=#]", nil)
		require.NoError(t, err)
		assert.True(t, result)

		// Empty values
		test(t, "#select1 option[value='']", []string{"option1a"}, document)
		// Empty values
		test(t, "#select1 option[value!='']", []string{"option1b", "option1c", "option1d"}, document)

		// Select options via :selected
		test(t, "#select1 option:selected", []string{"option1a"}, document)
		// Select options via :selected
		test(t, "#select2 option:selected", []string{"option2d"}, document)
		// Select options via :selected
		test(t, "#select3 option:selected", []string{"option3b", "option3c"}, document)
		// Select options via :selected
		test(t, "select[name='select2'] option:selected", []string{"option2d"}, document)

		// Grouped Form Elements
		test(t, "input[name='foo[bar]']", []string{"hidden2"}, document)

		input := domutils.GetElementById("text1", document, true)
		input.Attribs.Set("title", "Don't click me")

		// Quote within attribute value does not mess up tokenizer
		result, err = Is(input, "input[title=\"Don't click me\"]", nil)
		require.NoError(t, err)
		assert.True(t, result)

		// See jQuery #12303
		input.Attribs.Set("data-pos", ":first")
		// POS within attribute value is treated as an attribute value
		result, err = Is(input, "input[data-pos=\\:first]", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// POS within attribute value is treated as an attribute value
		result, err = Is(input, "input[data-pos=':first']", nil)
		require.NoError(t, err)
		assert.True(t, result)
		// POS within attribute value after pseudo is treated as an attribute value
		result, err = Is(input, ":input[data-pos=':first']", nil)
		require.NoError(t, err)
		assert.True(t, result)
		input.Attribs.Delete("data-pos")

		/*
		 * Make sure attribute value quoting works correctly. See jQuery #6093; #6428; #13894
		 * Use seeded results to bypass querySelectorAll optimizations
		 */
		attrbad := slices.Clone(parseDOM(
			`<input type='hidden' id='attrbad_space' name='foo bar'/>
				<input type='hidden' id='attrbad_dot' value='2' name='foo.baz'/>
				<input type='hidden' id='attrbad_brackets' value='2' name='foo[baz]'/>
				<input type='hidden' id='attrbad_injection' data-attr='foo_baz&#39;]'/>
				<input type='hidden' id='attrbad_quote' data-attr='&#39;'/>
				<input type='hidden' id='attrbad_backslash' data-attr='&#92;'/>
				<input type='hidden' id='attrbad_backslash_quote' data-attr='&#92;&#39;'/>
				<input type='hidden' id='attrbad_backslash_backslash' data-attr='&#92;&#92;'/>
				<input type='hidden' id='attrbad_unicode' data-attr='&#x4e00;'/>`,
			false,
		))

		for _, attr := range attrbad {
			domutils.AppendChild(
				domutils.GetElementById("qunit-fixture", document, true),
				attr,
			)
		}

		// Underscores don't need escaping
		test(t, "input[id=types_all]", []string{"types_all"}, document)

		// Escaped space
		test(t, "input[name=foo\\ bar]", []string{"attrbad_space"}, document)
		// Escaped dot
		test(t, "input[name=foo\\.baz]", []string{"attrbad_dot"}, document)
		// Escaped brackets
		test(t, "input[name=foo\\[baz\\]]", []string{"attrbad_brackets"}, document)

		// Escaped quote + right bracket
		test(t, "input[data-attr='foo_baz\\']']", []string{"attrbad_injection"}, document)

		// Quoted quote
		test(t, "input[data-attr='\\'']", []string{"attrbad_quote"}, document)
		// Quoted backslash
		test(t, "input[data-attr='\\\\']", []string{"attrbad_backslash"}, document)
		// Quoted backslash quote
		test(t, "input[data-attr='\\\\\\'']", []string{"attrbad_backslash_quote"}, document)
		// Quoted backslash backslash
		test(t, "input[data-attr='\\\\\\\\']", []string{"attrbad_backslash_backslash"}, document)

		// Quoted backslash backslash (numeric escape)
		test(t, "input[data-attr='\\5C\\\\']", []string{"attrbad_backslash_backslash"}, document)
		// Quoted backslash backslash (numeric escape with trailing space)
		test(t, "input[data-attr='\\5C \\\\']", []string{"attrbad_backslash_backslash"}, document)
		// Quoted backslash backslash (numeric escape with trailing tab)
		test(t, "input[data-attr='\\5C\t\\\\']", []string{"attrbad_backslash_backslash"}, document)
		// Long numeric escape (BMP)
		test(t, "input[data-attr='\\04e00']", []string{"attrbad_unicode"}, document)

		domutils.GetElementById("attrbad_unicode", document, true).Attribs.Set("data-attr", "\U0001D306A")
		/*
		 * It was too much code to fix Safari 5.x Supplemental Plane crashes (see ba5f09fa404379a87370ec905ffa47f8ac40aaa3)
		 * Long numeric escape (non-BMP)
		 */
		test(t, "input[data-attr='\\01D306A']", []string{"attrbad_unicode"}, document)

		for _, attr := range attrbad {
			domutils.RemoveElement(attr)
		}

		// `input[type=text]`
		test(t, "#form input[type=text]", []string{"text1", "text2", "hidden2", "name"}, document)
		// `input[type=search]`
		test(t, "#form input[type=search]", []string{"search"}, document)
		// `script[src]` (jQuery #13777)
		test(t, "#moretests script[src]", []string{"script-src"}, document)

		// #3279
		div := dom.NewElement("div", orderedmap.NewOrderedMap[string, string](), nil, "")
		div.Children = parseDOM("<div id='foo' xml:test='something'></div>", false)

		// Finding by attribute with escaped characters.
		children, err := SelectAll("[xml\\:test]", div.Children, nil)
		require.NoError(t, err)
		assert.Equal(t, children, div.Children)

		foo := domutils.GetElementById("foo", document, true)
		// Object.prototype property "constructor" (negative)',
		test(t, "[constructor]", []string{}, document)
		// Gecko Object.prototype property "watch" (negative)',
		test(t, "[watch]", []string{}, document)
		foo.Attribs.Set("constructor", "foo")
		foo.Attribs.Set("watch", "bar")
		// Object.prototype property "constructor"',
		test(t, "[constructor='foo']", []string{"foo"}, document)
		// Gecko Object.prototype property "watch"',
		test(t, "[watch='bar']", []string{"foo"}, document)

		// Value attribute is retrieved correctly
		test(t, "input[value=Test]", []string{"text1", "text2"}, document)
	})

	t.Run("pseudo - (parent|empty)", func(t *testing.T) {
		document := loadDoc()
		// Empty
		test(t, "ul:empty", []string{"firstUL"}, document)
		// Empty with comment node
		test(t, "ol:empty", []string{"empty"}, document)
		// Is A Parent
		test(t, "#qunit-fixture p:parent", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"sap",
			"first",
		}, document)
	})

	t.Run("pseudo - (first|last|only)-(child|of-type)", func(t *testing.T) {
		document := loadDoc()

		// First Child
		test(t, "p:first-child", []string{"firstp", "sndp"}, document)
		// First Child (leading id)
		test(t, "#qunit-fixture p:first-child", []string{"firstp", "sndp"}, document)
		// First Child (leading class)
		test(t, ".nothiddendiv div:first-child", []string{"nothiddendivchild"}, document)
		// First Child (case-insensitive)
		test(t, "#qunit-fixture p:FIRST-CHILD", []string{"firstp", "sndp"}, document)

		// Last Child
		test(t, "p:last-child", []string{"sap"}, document)
		// Last Child (leading id)
		test(t, "#qunit-fixture a:last-child", []string{
			"simon1",
			"anchor1",
			"mark",
			"yahoo",
			"anchor2",
			"simon",
			"liveLink1",
			"liveLink2",
		}, document)

		// Only Child
		test(t, "#qunit-fixture a:only-child", []string{
			"simon1",
			"anchor1",
			"yahoo",
			"anchor2",
			"liveLink1",
			"liveLink2",
		}, document)

		// First-of-type
		test(t, "#qunit-fixture > p:first-of-type", []string{"firstp"}, document)
		// Last-of-type
		test(t, "#qunit-fixture > p:last-of-type", []string{"first"}, document)
		// Only-of-type
		test(t, "#qunit-fixture > :only-of-type", []string{
			"name+value",
			"firstUL",
			"empty",
			"floatTest",
			"iframe",
			"table",
		}, document)

		// Verify that the child position isn't being cached improperly
		secondChildren, err := SelectAll("p:nth-child(2)", document, nil)
		require.NoError(t, err)

		newNodes := internal.MapFunc(secondChildren, func(child *dom.Node) *dom.Node {
			node := parseDOM("<div></div>", false)[0]
			domutils.Prepend(child, node)
			return node
		})

		// No longer second child
		test(t, "p:nth-child(2)", []string{}, document)

		for _, node := range newNodes {
			domutils.RemoveElement(node)
		}

		// Restored second child
		test(t, "p:nth-child(2)", []string{"ap", "en"}, document)
	})

	t.Run("pseudo - nth-child", func(t *testing.T) {
		document := loadDoc()

		// Nth-child
		test(t, "p:nth-child(1)", []string{"firstp", "sndp"}, document)
		// Nth-child (with whitespace)
		test(t, "p:nth-child( 1 )", []string{"firstp", "sndp"}, document)
		// Nth-child (case-insensitive)
		test(t, "#select1 option:NTH-child(3)", []string{"option1c"}, document)
		// Not nth-child
		test(t, "#qunit-fixture p:not(:nth-child(1))", []string{"ap", "en", "sap", "first"}, document)

		// Nth-child(2)
		test(t, "#qunit-fixture form#form > *:nth-child(2)", []string{"text1"}, document)
		// Nth-child(2)
		test(t, "#qunit-fixture form#form > :nth-child(2)", []string{"text1"}, document)

		// Nth-child(-1)
		test(t, "#select1 option:nth-child(-1)", []string{}, document)
		// Nth-child(3)
		test(t, "#select1 option:nth-child(3)", []string{"option1c"}, document)

		// "Nth-child(0n+3)"
		test(t, "#select1 option:nth-child(0n+3)", []string{"option1c"}, document)

		// Nth-child(1n+0)
		test(t, "#select1 option:nth-child(1n+0)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-child(1n)
		test(t, "#select1 option:nth-child(1n)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-child(n)
		test(t, "#select1 option:nth-child(n)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-child(even)
		test(t, "#select1 option:nth-child(even)", []string{"option1b", "option1d"}, document)
		// Nth-child(odd)
		test(t, "#select1 option:nth-child(odd)", []string{"option1a", "option1c"}, document)
		// Nth-child(2n)
		test(t, "#select1 option:nth-child(2n)", []string{"option1b", "option1d"}, document)
		// Nth-child(2n+1)
		test(t, "#select1 option:nth-child(2n+1)", []string{"option1a", "option1c"}, document)
		// Nth-child(2n + 1)
		test(t, "#select1 option:nth-child(2n + 1)", []string{"option1a", "option1c"}, document)
		// Nth-child(+2n + 1)
		test(t, "#select1 option:nth-child(+2n + 1)", []string{"option1a", "option1c"}, document)
		// Nth-child(3n)
		test(t, "#select1 option:nth-child(3n)", []string{"option1c"}, document)
		// Nth-child(3n+1)
		test(t, "#select1 option:nth-child(3n+1)", []string{"option1a", "option1d"}, document)
		// Nth-child(3n+2)
		test(t, "#select1 option:nth-child(3n+2)", []string{"option1b"}, document)
		// Nth-child(3n+3)
		test(t, "#select1 option:nth-child(3n+3)", []string{"option1c"}, document)
		// Nth-child(3n-1)
		test(t, "#select1 option:nth-child(3n-1)", []string{"option1b"}, document)
		// Nth-child(3n-2)
		test(t, "#select1 option:nth-child(3n-2)", []string{"option1a", "option1d"}, document)
		// Nth-child(3n-3)
		test(t, "#select1 option:nth-child(3n-3)", []string{"option1c"}, document)
		// Nth-child(3n+0)
		test(t, "#select1 option:nth-child(3n+0)", []string{"option1c"}, document)
		// Nth-child(-1n+3)
		test(t, "#select1 option:nth-child(-1n+3)", []string{
			"option1a",
			"option1b",
			"option1c",
		}, document)
		// Nth-child(-n+3)
		test(t, "#select1 option:nth-child(-n+3)", []string{
			"option1a",
			"option1b",
			"option1c",
		}, document)
		// Nth-child(-1n + 3)
		test(t, "#select1 option:nth-child(-1n + 3)", []string{
			"option1a",
			"option1b",
			"option1c",
		}, document)
	})

	t.Run("pseudo - nth-last-child", func(t *testing.T) {
		document := loadDoc()

		// Nth-last-child
		test(t, "form:nth-last-child(5)", []string{"testForm"}, document)
		// Nth-last-child (with whitespace)
		test(t, "form:nth-last-child( 5 )", []string{"testForm"}, document)
		// Nth-last-child (case-insensitive)
		test(t, "#select1 option:NTH-last-child(3)", []string{"option1b"}, document)
		// Not nth-last-child
		test(t, "#qunit-fixture p:not(:nth-last-child(1))", []string{
			"firstp",
			"ap",
			"sndp",
			"en",
			"first",
		}, document)

		// Nth-last-child(-1)
		test(t, "#select1 option:nth-last-child(-1)", []string{}, document)
		// Nth-last-child(3)
		test(t, "#select1 :nth-last-child(3)", []string{"option1b"}, document)
		// Nth-last-child(3)
		test(t, "#select1 *:nth-last-child(3)", []string{"option1b"}, document)
		// Nth-last-child(3)
		test(t, "#select1 option:nth-last-child(3)", []string{"option1b"}, document)

		// Nth-last-child(0n+3)
		test(t, "#select1 option:nth-last-child(0n+3)", []string{"option1b"}, document)

		// Nth-last-child(1n+0)
		test(t, "#select1 option:nth-last-child(1n+0)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-last-child(1n)
		test(t, "#select1 option:nth-last-child(1n)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-last-child(n)
		test(t, "#select1 option:nth-last-child(n)", []string{
			"option1a",
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-last-child(even)
		test(t, "#select1 option:nth-last-child(even)", []string{"option1a", "option1c"}, document)
		// Nth-last-child(odd)
		test(t, "#select1 option:nth-last-child(odd)", []string{"option1b", "option1d"}, document)
		// Nth-last-child(2n)
		test(t, "#select1 option:nth-last-child(2n)", []string{"option1a", "option1c"}, document)
		// Nth-last-child(2n+1)
		test(t, "#select1 option:nth-last-child(2n+1)", []string{"option1b", "option1d"}, document)
		// Nth-last-child(2n + 1)
		test(t, "#select1 option:nth-last-child(2n + 1)", []string{"option1b", "option1d"}, document)
		// Nth-last-child(+2n + 1)
		test(t, "#select1 option:nth-last-child(+2n + 1)", []string{"option1b", "option1d"}, document)
		// Nth-last-child(3n)
		test(t, "#select1 option:nth-last-child(3n)", []string{"option1b"}, document)
		// Nth-last-child(3n+1)
		test(t, "#select1 option:nth-last-child(3n+1)", []string{"option1a", "option1d"}, document)
		// Nth-last-child(3n+2)
		test(t, "#select1 option:nth-last-child(3n+2)", []string{"option1c"}, document)
		// Nth-last-child(3n+3)
		test(t, "#select1 option:nth-last-child(3n+3)", []string{"option1b"}, document)
		// Nth-last-child(3n-1)
		test(t, "#select1 option:nth-last-child(3n-1)", []string{"option1c"}, document)
		// Nth-last-child(3n-2)
		test(t, "#select1 option:nth-last-child(3n-2)", []string{"option1a", "option1d"}, document)
		// Nth-last-child(3n-3)
		test(t, "#select1 option:nth-last-child(3n-3)", []string{"option1b"}, document)
		// Nth-last-child(3n+0)
		test(t, "#select1 option:nth-last-child(3n+0)", []string{"option1b"}, document)
		// Nth-last-child(-1n+3)
		test(t, "#select1 option:nth-last-child(-1n+3)", []string{
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-last-child(-n+3)
		test(t, "#select1 option:nth-last-child(-n+3)", []string{
			"option1b",
			"option1c",
			"option1d",
		}, document)
		// Nth-last-child(-1n + 3)
		test(t, "#select1 option:nth-last-child(-1n + 3)", []string{
			"option1b",
			"option1c",
			"option1d",
		}, document)
	})

	t.Run("pseudo - nth-of-type", func(t *testing.T) {
		document := loadDoc()
		// Nth-of-type(-1)
		test(t, ":nth-of-type(-1)", []string{}, document)
		// Nth-of-type(3)
		test(t, "#ap :nth-of-type(3)", []string{"mark"}, document)
		// Nth-of-type(n)
		test(t, "#ap :nth-of-type(n)", []string{
			"google",
			"groups",
			"code1",
			"anchor1",
			"mark",
		}, document)
		// Nth-of-type(0n+3)
		test(t, "#ap :nth-of-type(0n+3)", []string{"mark"}, document)
		// Nth-of-type(2n)
		test(t, "#ap :nth-of-type(2n)", []string{"groups"}, document)
		// Nth-of-type(even)
		test(t, "#ap :nth-of-type(even)", []string{"groups"}, document)
		// Nth-of-type(2n+1)
		test(t, "#ap :nth-of-type(2n+1)", []string{"google", "code1", "anchor1", "mark"}, document)
		// Nth-of-type(odd)
		test(t, "#ap :nth-of-type(odd)", []string{"google", "code1", "anchor1", "mark"}, document)
		// Nth-of-type(-n+2)
		test(t, "#qunit-fixture > :nth-of-type(-n+2)", []string{
			"firstp",
			"ap",
			"foo",
			"nothiddendiv",
			"name+value",
			"firstUL",
			"empty",
			"form",
			"floatTest",
			"iframe",
			"lengthtest",
			"table",
		}, document)
	})

	t.Run("pseudo - nth-last-of-type", func(t *testing.T) {
		document := loadDoc()
		// Nth-last-of-type(-1)
		test(t, ":nth-last-of-type(-1)", []string{}, document)
		// Nth-last-of-type(3)
		test(t, "#ap :nth-last-of-type(3)", []string{"google"}, document)
		// Nth-last-of-type(n)
		test(t, "#ap :nth-last-of-type(n)", []string{
			"google",
			"groups",
			"code1",
			"anchor1",
			"mark",
		}, document)
		// Nth-last-of-type(0n+3)
		test(t, "#ap :nth-last-of-type(0n+3)", []string{"google"}, document)
		// Nth-last-of-type(2n)
		test(t, "#ap :nth-last-of-type(2n)", []string{"groups"}, document)
		// Nth-last-of-type(even)
		test(t, "#ap :nth-last-of-type(even)", []string{"groups"}, document)
		// Nth-last-of-type(2n+1)
		test(t, "#ap :nth-last-of-type(2n+1)", []string{
			"google",
			"code1",
			"anchor1",
			"mark",
		}, document)
		// Nth-last-of-type(odd)
		test(t, "#ap :nth-last-of-type(odd)", []string{"google", "code1", "anchor1", "mark"}, document)
		// Nth-last-of-type(-n+2)
		test(t, "#qunit-fixture > :nth-last-of-type(-n+2)", []string{
			"ap",
			"name+value",
			"first",
			"firstUL",
			"empty",
			"floatTest",
			"iframe",
			"table",
			"name-tests",
			"testForm",
			"liveHandlerOrder",
			"siblingTest",
		}, document)
	})

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

	t.Run("regression - has", func(t *testing.T) {
		html := `<div id="outer">
			<section id="target">
				<p id="inner"></p>
			</section>
		</div>`

		doc, err := dom.ParseDocument(html, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
		require.NoError(t, err)

		matches, err := SelectAll("section:has(div p)", doc.Children, nil)
		require.NoError(t, err)

		assert.Len(t, matches, 0)
	})

	t.Run("pseudo - form", func(t *testing.T) {
		document := loadDoc()

		extraTexts, err := dom.ParseDocument(
			`<input id="impliedText"/><input id="capitalText" type="TEXT">`,
			&parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true},
		)
		require.NoError(t, err)

		form := domutils.GetElementById("form", document, true)

		for _, text := range slices.Clone(extraTexts.Children) {
			domutils.AppendChild(form, text)
		}

		// Form element :input
		test(t, "#form :input", []string{
			"text1",
			"text2",
			"radio1",
			"radio2",
			"check1",
			"check2",
			"hidden1",
			"hidden2",
			"name",
			"search",
			"button",
			"area1",
			"select1",
			"select2",
			"select3",
			"select4",
			"select5",
			"impliedText",
			"capitalText",
		}, document)
		// Form element :radio
		test(t, "#form :radio", []string{"radio1", "radio2"}, document)
		// Form element :checkbox
		test(t, "#form :checkbox", []string{"check1", "check2"}, document)
		// Form element :text
		test(t, "#form :text", []string{
			"text1",
			"text2",
			"hidden2",
			"name",
			"impliedText",
			"capitalText",
		}, document)
		// Form element :radio:checked
		test(t, "#form :radio:checked", []string{"radio2"}, document)
		// Form element :checkbox:checked
		test(t, "#form :checkbox:checked", []string{"check1"}, document)
		// Form element :radio:checked, :checkbox:checked
		test(t, "#form :radio:checked, #form :checkbox:checked", []string{
			"radio2",
			"check1",
		}, document)

		// Selected Option Element
		test(t, "#form option:selected", []string{
			"option1a",
			"option2d",
			"option3b",
			"option3c",
			"option4b",
			"option4c",
			"option4d",
			"option5a",
		}, document)
		// Selected Option Element are also :checked
		test(t, "#form option:checked", []string{
			"option1a",
			"option2d",
			"option3b",
			"option3c",
			"option4b",
			"option4c",
			"option4d",
			"option5a",
		}, document)
		// Hidden inputs should be treated as enabled. See QSA test.
		test(t, "#hidden1:enabled", []string{"hidden1"}, document)
	})

	t.Run("pseudo - :root", func(t *testing.T) {
		document := loadDoc()
		idx := slices.IndexFunc(document, func(n *dom.Node) bool {
			return dom.IsTag(n)
		})
		require.Positive(t, idx)
		root := document[idx]
		// :root selector
		found, err := SelectOne(":root", document, nil)
		require.NoError(t, err)

		assert.Equal(t, root, found)
	})

	t.Run("caching", func(t *testing.T) {
		document := loadDoc()
		ap := domutils.GetElementById("ap", document, true)
		_, err := SelectAll(":not(code)", []*dom.Node{ap}, nil)
		require.NoError(t, err)
		// Reusing selector with new context
		foo := domutils.GetElementById("foo", document, true)
		require.NotNil(t, foo)
		test(t,
			":not(code)",
			[]string{"sndp", "en", "yahoo", "sap", "anchor2", "simon"},
			foo.Children,
		)
	})

	t.Run("more specific selector should find less elements", func(t *testing.T) {
		document := loadDoc()
		// Same selector, but with an attribute filter added
		one, err := SelectAll("#qunit-fixture div div", document, nil)
		require.NoError(t, err)
		another, err := SelectAll("#qunit-fixture div div[id]", document, nil)
		require.NoError(t, err)
		assert.Greater(t, len(one), len(another))
	})
}
