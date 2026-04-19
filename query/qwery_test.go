package query

import (
	_ "embed"
	"math"
	"testing"

	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
	"github.com/stretchr/testify/assert"
)

//go:embed __fixtures__/qwery.html
var qwery string

func clearNode(node *dom.Node) *dom.Node {
	node.Parent = nil
	node.NextSibling = nil
	node.PreviousSibling = nil
	for _, child := range node.Children {
		clearNode(child)
	}
	return node
}

var hash = ""

// const location = { hash: "" };
var options = &types.Options{
	Pseudos: map[string]func(elem *dom.Node, value string) bool{
		"target": func(elem *dom.Node, subselect string) bool {
			var value string
			if len(hash) > 1 {
				value = hash[1:]
			}
			return hash != "" && domutils.GetAttributeValue(elem, "id") == value
		},
	},
}

func selectAll(selector string, context []*dom.Node) []*dom.Node {
	if context == nil {
		node := parseDocument(qwery)
		context = node.Children
	}

	matches, err := SelectAll(selector, context, options)
	if err != nil {
		panic(err)
	}

	return matches
}

func _selectAll(selector string, context []*dom.Node) []*dom.Node {
	return internal.MapFunc(selectAll(selector, context), clearNode)
}

func getElementById(id string, context []*dom.Node) *dom.Node {
	if context == nil {
		node := parseDocument(qwery)
		context = node.Children
	}

	return clearNode(domutils.GetElementById(id, context, true))
}

/*
 * Adapted from https://github.com/ded/qwery/blob/master/tests/tests.js
 */

var frag = parseDOM(
	`<root><div class="d i v">`+
		`<p id="oooo"><em></em><em id="emem"></em></p>`+
		"</div>"+
		`<p id="sep">`+
		`<div class="a"><span></span></div>`+
		"</p></root>",
	false,
)

var doc = parseDOM(
	`<root><div id="hsoob">`+
		`<div class="a b">`+
		`<div class="d e sib" test="fg" id="booshTest"><p><span id="spanny"></span></p></div>`+
		`<em nopass="copyrighters" rel="copyright booshrs" test="f g" class="sib"></em>`+
		`<span class="h i a sib"></span>`+
		"</div>"+
		`<p class="odd"></p>`+
		"</div>"+
		`<div id="lonelyHsoob"></div></root>`,
	false,
)

//var el = getElementById("attr-child-boosh", nil)

var pseudos = internal.FilterFunc(
	getElementById("pseudos", nil).Children,
	func(n *dom.Node) bool { return dom.IsTag(n) },
)

func TestQwery(t *testing.T) {
	t.Run("Contexts", func(t *testing.T) {
		t.Run("should be able to pass optional context", func(t *testing.T) {
			assert.Len(t, selectAll(".a", nil), 3)                      // No context found 3 elements (.a)
			assert.Len(t, selectAll(".a", selectAll("#boosh", nil)), 2) // Context found 2 elements (#boosh .a)
		})

		t.Run("should be able to pass qwery result as context", func(t *testing.T) {
			assert.Len(t, selectAll(".a", selectAll("#boosh", nil)), 2)      // Context found 2 elements(.a, #boosh)
			assert.Len(t, selectAll("> .a", selectAll(".a", nil)), 1)        // Context found 0 elements(.a, .a)
			assert.Len(t, selectAll("> .a", selectAll(".b", nil)), 1)        // Context found 1 elements(.a, .b)
			assert.Len(t, selectAll("> .a", selectAll("#boosh .b", nil)), 1) // Context found 1 elements(.a, #boosh .b)
			assert.Len(t, selectAll("> .b", selectAll("#boosh .b", nil)), 0) // Context found 0 elements(.b, #boosh .b)
		})

		t.Run("should not return duplicates from combinators", func(t *testing.T) {
			assert.Len(t, selectAll("#boosh,#boosh", nil), 1)         // Two booshes dont make a thing go right
			assert.Len(t, selectAll("#boosh,.apples,#boosh", nil), 1) // Two booshes and an apple dont make a thing go right
		})

		t.Run("byId sub-queries within context", func(t *testing.T) {
			assert.Len(t, selectAll("#booshTest", selectAll("#boosh", nil)), 1)      // Found "#id #id"
			assert.Len(t, selectAll(".a.b #booshTest", selectAll("#boosh", nil)), 1) // Found ".class.class #id"
			assert.Len(t, selectAll(".a>#booshTest", selectAll("#boosh", nil)), 1)   // Found ".class>#id"
			assert.Len(t, selectAll(">.a>#booshTest", selectAll("#boosh", nil)), 1)  // Found ">.class>#id"
			assert.Len(t, selectAll("#boosh", selectAll("#booshTest", nil)), 0)      // Shouldn't find #boosh (ancestor) within #booshTest (descendent)
			assert.Len(t, selectAll("#boosh", selectAll("#lonelyBoosh", nil)), 0)    // Shouldn't find #boosh within #lonelyBoosh (unrelated)
		})
	})

	t.Run("CSS 1", func(t *testing.T) {
		t.Run("get element by id", func(t *testing.T) {
			result := selectAll("#boosh", nil)
			assert.NotNil(t, result[0])               // Found element with id=boosh
			assert.NotNil(t, selectAll("h1", nil)[0]) // Found 1 h1
		})

		t.Run("byId sub-queries", func(t *testing.T) {
			assert.Len(t, selectAll("#boosh #booshTest", nil), 1)    // Found "#id #id"
			assert.Len(t, selectAll(".a.b #booshTest", nil), 1)      // Found ".class.class #id"
			assert.Len(t, selectAll("#boosh>.a>#booshTest", nil), 1) // Found "#id>.class>#id"
			assert.Len(t, selectAll(".a>#booshTest", nil), 1)        // Found ".class>#id"
		})

		t.Run("get elements by class", func(t *testing.T) {
			assert.Len(t, selectAll("#boosh .a", nil), 2)         // Found two elements
			assert.NotNil(t, selectAll("#boosh div.a", nil)[0])   // Found one element
			assert.Len(t, selectAll("#boosh div", nil), 2)        // Found two {div} elements
			assert.NotNil(t, selectAll("#boosh span", nil)[0])    // Found one {span} element
			assert.NotNil(t, selectAll("#boosh div div", nil)[0]) // Found a single div
			assert.Len(t, selectAll("a.odd", nil), 1)             // Found single a
		})

		t.Run("combos", func(t *testing.T) {
			assert.Len(t, selectAll("#boosh div,#boosh span", nil), 3) // Found 2 divs and 1 span
		})

		t.Run("class with dashes", func(t *testing.T) {
			assert.Len(t, selectAll(".class-with-dashes", nil), 1) // Found something
		})

		t.Run("should ignore comment nodes", func(t *testing.T) {
			assert.Len(t, selectAll("#boosh *", nil), 4) // Found only 4 elements under #boosh
		})

		t.Run("deep messy relationships", func(t *testing.T) {
			/*
			 * These are mostly characterised by a combination of tight relationships and loose relationships
			 * on the right side of the query it's easy to find matches but they tighten up quickly as you
			 * go to the left
			 * they are useful for making sure the dom crawler doesn't stop short or over-extend as it works
			 * up the tree the crawl needs to be comprehensive
			 */
			assert.Len(t, selectAll("div#fixtures > div a", nil), 5)                                        // Found four results for "div#fixtures > div a"
			assert.Len(t, selectAll(".direct-descend > .direct-descend .lvl2", nil), 1)                     // Found one result for ".direct-descend > .direct-descend .lvl2"
			assert.Len(t, selectAll(".direct-descend > .direct-descend div", nil), 1)                       // Found one result for ".direct-descend > .direct-descend div"
			assert.Len(t, selectAll(".direct-descend > .direct-descend div", nil), 1)                       // Found one result for ".direct-descend > .direct-descend div"
			assert.Len(t, selectAll("div#fixtures div ~ a div", nil), 0)                                    // Found no results for odd query
			assert.Len(t, selectAll(".direct-descend > .direct-descend > .direct-descend ~ .lvl2", nil), 0) // Found no results for another odd query
		})
	})

	t.Run("CSS 2", func(t *testing.T) {
		t.Run("get elements by attribute", func(t *testing.T) {
			wanted := _selectAll("#boosh div[test]", nil)[0]
			expected := getElementById("booshTest", nil)
			assert.Equal(t, expected, wanted)                                    // Found attribute
			assert.Equal(t, expected, _selectAll("#boosh div[test=fg]", nil)[0]) // Found attribute with value
			assert.Len(t, selectAll(`em[rel~="copyright"]`, nil), 1)             // Found em[rel~="copyright"]
			assert.Len(t, selectAll(`em[nopass~="copyright"]`, nil), 0)          // Found em[nopass~="copyright"]
		})

		t.Run("should not throw error by attribute selector", func(t *testing.T) {
			assert.Len(t, selectAll(`[foo^="bar"]`, nil), 1) // Found 1 element
		})

		t.Run("crazy town", func(t *testing.T) {
			el := getElementById("attr-test3", nil)
			assert.Equal(t, el, _selectAll(`div#attr-test3.found.you[title="whatup duders"]`, nil)[0]) // Found the right element
		})
	})

	t.Run("attribute selectors", func(t *testing.T) {
		/* CSS 2 SPEC */

		t.Run("[attr]", func(t *testing.T) {
			expected := getElementById("attr-test-1", nil)
			assert.Equal(t, expected, _selectAll("#attributes div[unique-test]", nil)[0]) // Found attribute with [attr]
		})

		t.Run("[attr=val]", func(t *testing.T) {
			expected := getElementById("attr-test-2", nil)
			assert.Equal(t, expected, _selectAll(`#attributes div[test="two-foo"]`, nil)[0]) // Found attribute with =
			assert.Equal(t, expected, _selectAll("#attributes div[test='two-foo']", nil)[0]) // Found attribute with =
			assert.Equal(t, expected, _selectAll("#attributes div[test=two-foo]", nil)[0])   // Found attribute with =
		})

		t.Run("[attr~=val]", func(t *testing.T) {
			expected := getElementById("attr-test-3", nil)
			assert.Equal(t, expected, _selectAll("#attributes div[test~=three]", nil)[0]) // Found attribute with ~=
		})

		t.Run("[attr|=val]", func(t *testing.T) {
			expected := getElementById("attr-test-2", nil)
			assert.Equal(t, expected, _selectAll(`#attributes div[test|="two-foo"]`, nil)[0]) // Found attribute with |=
			assert.Equal(t, expected, _selectAll("#attributes div[test|=two]", nil)[0])       // Found attribute with |=
		})

		t.Run("[href=#x] special case", func(t *testing.T) {
			expected := getElementById("attr-test-4", nil)
			assert.Equal(t, expected, _selectAll(`#attributes a[href="#aname"]`, nil)[0]) // Found attribute with href=#x
		})

		/* CSS 3 SPEC */
		t.Run("[attr^=val]", func(t *testing.T) {
			expected := getElementById("attr-test-2", nil)
			assert.Equal(t, expected, _selectAll("#attributes div[test^=two]", nil)[0]) // Found attribute with ^=
		})

		t.Run("[attr$=val]", func(t *testing.T) {
			expected := getElementById("attr-test-2", nil)
			assert.Equal(t, expected, _selectAll("#attributes div[test$=foo]", nil)[0]) // Found attribute with $=
		})

		t.Run("[attr*=val]", func(t *testing.T) {
			expected := getElementById("attr-test-3", nil)
			assert.Equal(t, expected, _selectAll("#attributes div[test*=hree]", nil)[0]) // Found attribute with *=
		})

		t.Run("direct descendants", func(t *testing.T) {
			assert.Len(t, selectAll("#direct-descend > .direct-descend", nil), 2)         // Found two direct descendents
			assert.Len(t, selectAll("#direct-descend > .direct-descend > .lvl2", nil), 3) // Found three second-level direct descendents
		})

		t.Run("sibling elements", func(t *testing.T) {
			assert.Len(t, selectAll("#sibling-selector ~ .sibling-selector", nil), 2)    // Found two siblings
			assert.Len(t, selectAll("#sibling-selector ~ div.sibling-selector", nil), 2) // Found two siblings
			assert.Len(t, selectAll("#sibling-selector + div.sibling-selector", nil), 1) // Found one sibling
			assert.Len(t, selectAll("#sibling-selector + .sibling-selector", nil), 1)    // Found one sibling

			assert.Len(t, selectAll(".parent .oldest ~ .sibling", nil), 4)   // Found four younger siblings
			assert.Len(t, selectAll(".parent .middle ~ .sibling", nil), 2)   // Found two younger siblings
			assert.Len(t, selectAll(".parent .middle ~ h4", nil), 1)         // Found next sibling by tag
			assert.Len(t, selectAll(".parent .middle ~ h4.younger", nil), 1) // Found next sibling by tag and class
			assert.Len(t, selectAll(".parent .middle ~ h3", nil), 0)         // An element can't be its own sibling
			assert.Len(t, selectAll(".parent .middle ~ h2", nil), 0)         // Didn't find an older sibling
			assert.Len(t, selectAll(".parent .youngest ~ .sibling", nil), 0) // Found no younger siblings

			assert.Len(t, selectAll(".parent .oldest + .sibling", nil), 1)   // Found next sibling
			assert.Len(t, selectAll(".parent .middle + .sibling", nil), 1)   // Found next sibling
			assert.Len(t, selectAll(".parent .middle + h4", nil), 1)         // Found next sibling by tag
			assert.Len(t, selectAll(".parent .middle + h3", nil), 0)         // An element can't be its own sibling
			assert.Len(t, selectAll(".parent .middle + h2", nil), 0)         // Didn't find an older sibling
			assert.Len(t, selectAll(".parent .youngest + .sibling", nil), 0) // Found no younger siblings
		})
	})

	t.Run("element-context queries", func(t *testing.T) {
		t.Run("relationship-first queries", func(t *testing.T) {
			assert.Len(t, selectAll("> .direct-descend", selectAll("#direct-descend", nil)), 2)     // Found two direct descendents using > first
			assert.Len(t, selectAll("~ .sibling-selector", selectAll("#sibling-selector", nil)), 2) // Found two siblings with ~ first
			assert.Len(t, selectAll("+ .sibling-selector", selectAll("#sibling-selector", nil)), 1) // Found one sibling with + first
			assert.Len(t, selectAll("> .tokens a", []*dom.Node{selectAll(".idless", nil)[0]}), 1)   // Found one sibling from a root with no id
		})

		// Should be able to query on an element that hasn't been inserted into the dom
		t.Run("detached fragments", func(t *testing.T) {
			assert.Len(t, selectAll(".a span", frag), 1)    // Should find child elements of fragment
			assert.Len(t, selectAll("> div p em", frag), 2) // Should find child elements of fragment, relationship first
		})

		t.Run("byId sub-queries within detached fragment", func(t *testing.T) {
			assert.Len(t, selectAll("#emem", frag), 1)                     // Found "#id" in fragment
			assert.Len(t, selectAll(".d.i #emem", frag), 1)                // Found ".class.class #id" in fragment
			assert.Len(t, selectAll(".d #oooo #emem", frag), 1)            // Found ".class #id #id" in fragment
			assert.Len(t, selectAll("> div #oooo", frag), 1)               // Found "> .class #id" in fragment
			assert.Len(t, selectAll("#oooo", selectAll("#emem", frag)), 0) // Shouldn't find #oooo (ancestor) within #emem (descendent)
			assert.Len(t, selectAll("#sep", selectAll("#emem", frag)), 0)  // Shouldn't find #sep within #emem (unrelated)
		})

		t.Run("exclude self in match", func(t *testing.T) {
			assert.Len(t, selectAll(".order-matters", selectAll("#order-matters", nil)[0].Children), 4) // Should not include self in element-context queries
		})

		// Because form's have .length
		t.Run("forms can be used as contexts", func(t *testing.T) {
			assert.Len(t, selectAll("*", selectAll("form", nil)[0].Children), 3) // Found 3 elements under &lt;form&gt;
		})
	})

	t.Run("tokenizer", func(t *testing.T) {
		t.Run("should not get weird tokens", func(t *testing.T) {
			assert.Equal(t,
				_selectAll(`div .tokens[title="one"]`, nil)[0],
				getElementById("token-one", nil),
			) // Found div .tokens[title="one"]
			assert.Equal(t,
				_selectAll(`div .tokens[title="one two"]`, nil)[0],
				getElementById("token-two", nil),
			) // Found div .tokens[title="one two"]
			assert.Equal(t,
				_selectAll(`div .tokens[title="one two three #%"]`, nil)[0],
				getElementById("token-three", nil),
			) // Found div .tokens[title="one two three #%"]
			assert.Equal(t,
				_selectAll("div .tokens[title='one two three #%'] a", nil)[0],
				getElementById("token-four", nil),
			) // Found div .tokens[title=\'one two three #%\'] a
			assert.Equal(t,
				_selectAll(`div .tokens[title="one two three #%"] a[href$=foo] div`, nil)[0],
				getElementById("token-five", nil),
			) // Found div .tokens[title="one two three #%"] a[href=foo] div
		})
	})

	t.Run("interesting syntaxes", func(t *testing.T) {
		t.Run("should parse bad selectors", func(t *testing.T) {
			assert.Greater(t, len(selectAll("#spaced-tokens    p    em    a", nil)), 0) // Found element with funny tokens
		})
	})

	t.Run("order matters", func(t *testing.T) {
		/*
		 * <div id="order-matters">
		 *   <p class="order-matters"></p>
		 *   <a class="order-matters">
		 *     <em class="order-matters"></em><b class="order-matters"></b>
		 *   </a>
		 * </div>
		 */

		t.Run("the order of elements return matters", func(t *testing.T) {
			els := selectAll("#order-matters .order-matters", nil)
			assert.Equal(t, els[0].Name, "p")  // First element matched is a {p} tag
			assert.Equal(t, els[1].Name, "a")  // First element matched is a {a} tag
			assert.Equal(t, els[2].Name, "em") // First element matched is a {em} tag
			assert.Equal(t, els[3].Name, "b")  // First element matched is a {b} tag
		})
	})

	t.Run("pseudo-selectors", func(t *testing.T) {
		t.Run(":contains", func(t *testing.T) {
			assert.Len(t, selectAll("li:contains(humans)", nil), 1) // Found by "element:contains(text)"
			assert.Len(t, selectAll(":contains(humans)", nil), 5)   // Found by ":contains(text)", including all ancestors
			// * Is an important case, can cause weird errors
			assert.Len(t, selectAll("*:contains(humans)", nil), 5)  // Found by "*:contains(text)", including all ancestors
			assert.Len(t, selectAll("ol:contains(humans)", nil), 1) // Found by "ancestor:contains(text)"
		})

		t.Run(":not", func(t *testing.T) {
			assert.Len(t, selectAll(".odd:not(div)", nil), 1) // Found one .odd :not an &lt;a&gt;
		})

		t.Run(":first-child", func(t *testing.T) {
			assert.Equal(t, _selectAll("#pseudos div:first-child", nil)[0], pseudos[0]) // Found first child
			assert.Len(t, selectAll("#pseudos div:first-child", nil), 1)                // Found only 1
		})

		t.Run(":last-child", func(t *testing.T) {
			all := domutils.GetElementsByTagName("div", pseudos, false, math.MaxInt)
			assert.Equal(t, _selectAll("#pseudos div:last-child", nil)[0], all[len(all)-1]) // Found last child
			assert.Len(t, selectAll("#pseudos div:last-child", nil), 1)                     // Found only 1
		})

		t.Run(`ol > li[attr="boosh"]:last-child`, func(t *testing.T) {
			expected := getElementById("attr-child-boosh", nil)
			assert.Len(t, selectAll(`ol > li[attr="boosh"]:last-child`, nil), 1)              // Only 1 element found
			assert.Equal(t, _selectAll(`ol > li[attr="boosh"]:last-child`, nil)[0], expected) // Found correct element
		})

		t.Run(":first-of-type", func(t *testing.T) {
			assert.Equal(t,
				_selectAll("#pseudos a:first-of-type", nil)[0],
				domutils.GetElementsByTagName("a", pseudos, true, math.MaxInt)[0],
			) // Found first a element
			assert.Len(t, selectAll("#pseudos a:first-of-type", nil), 1) // Found only 1
		})

		t.Run(":last-of-type", func(t *testing.T) {
			all := domutils.GetElementsByTagName("div", pseudos, true, math.MaxInt)
			assert.Equal(t, _selectAll("#pseudos div:last-of-type", nil)[0], all[len(all)-1]) // Found last div element
			assert.Len(t, selectAll("#pseudos div:last-of-type", nil), 1)                     // Found only 1
		})

		t.Run(":only-of-type", func(t *testing.T) {
			assert.Equal(t,
				_selectAll("#pseudos a:only-of-type", nil)[0],
				domutils.GetElementsByTagName("a", pseudos, true, math.MaxInt)[0],
			) // Found the only a element
			assert.Len(t, selectAll("#pseudos a:first-of-type", nil), 1) // Found only 1
		})

		t.Run(":target", func(t *testing.T) {
			hash = ""
			assert.Len(t, selectAll("#pseudos:target", nil), 0) // #pseudos is not the target
			hash = "#pseudos"
			assert.Len(t, selectAll("#pseudos:target", nil), 1) // Now #pseudos is the target
			hash = ""
		})
	})

	t.Run("selecting elements in other documents", func(t *testing.T) {
		t.Run("get element by id", func(t *testing.T) {
			result := selectAll("#hsoob", doc)
			assert.NotNil(t, result[0]) // Found element with id=hsoob
		})

		t.Run("get elements by class", func(t *testing.T) {
			assert.Len(t, selectAll("#hsoob .a", doc), 2)         // Found two elements
			assert.NotNil(t, selectAll("#hsoob div.a", doc)[0])   // Found one element
			assert.Len(t, selectAll("#hsoob div", doc), 2)        // Found two {div} elements
			assert.NotNil(t, selectAll("#hsoob span", doc)[0])    // Found one {span} element
			assert.NotNil(t, selectAll("#hsoob div div", doc)[0]) // Found a single div
			assert.Len(t, selectAll("p.odd", doc), 1)             // Found single br
		})

		t.Run("complex selectors", func(t *testing.T) {
			assert.Len(t, selectAll(".d ~ .sib", doc), 2)                // Found one ~ sibling
			assert.Len(t, selectAll(".a .d + .sib", doc), 1)             // Found 2 + siblings
			assert.Len(t, selectAll("#hsoob > div > .h", doc), 1)        // Found span using child selectors
			assert.Len(t, selectAll(`.a .d ~ .sib[test="f g"]`, doc), 1) // Found 1 ~ sibling with test attribute
		})

		t.Run("byId sub-queries", func(t *testing.T) {
			assert.Len(t, selectAll("#hsoob #spanny", doc), 1)        // Found "#id #id" in frame
			assert.Len(t, selectAll(".a #spanny", doc), 1)            // Found ".class #id" in frame
			assert.Len(t, selectAll(".a #booshTest #spanny", doc), 1) // Found ".class #id #id" in frame
			assert.Len(t, selectAll("> #hsoob", doc), 1)              // Found "> #id" in frame
		})

		t.Run("byId sub-queries within sub-context", func(t *testing.T) {
			assert.Len(t, selectAll("#spanny", selectAll("#hsoob", doc)), 1)               // Found "#id -> #id" in frame
			assert.Len(t, selectAll(".a #spanny", selectAll("#hsoob", doc)), 1)            // Found ".class #id" in frame
			assert.Len(t, selectAll(".a #booshTest #spanny", selectAll("#hsoob", doc)), 1) // Found ".class #id #id" in frame
			assert.Len(t, selectAll(".a > #booshTest", selectAll("#hsoob", doc)), 1)       // Found "> .class #id" in frame
			assert.Len(t, selectAll("#booshTest", selectAll("#spanny", doc)), 0)           // Shouldn't find #booshTest (ancestor) within #spanny (descendent)
			assert.Len(t, selectAll("#booshTest", selectAll("#lonelyHsoob", doc)), 0)      // Shouldn't find #booshTest within #lonelyHsoob (unrelated)
		})
	})
}
