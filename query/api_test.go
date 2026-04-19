package query

import (
	"strings"
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func parseDOM(str string, xmlMode bool) []*dom.Node {
	return dom.ParseDOM(str, &parser.ParserOptions{XmlMode: xmlMode, LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
}

func TestApi(t *testing.T) {
	doc := parseDOM("<div id=foo><p>foo</p></div>", false)[0]
	//xmlDom := parseDOM("<DiV id=foo><P>foo</P></DiV>", true)[0]
	notYet := "not yet supported by css-select"

	t.Run("removes duplicates", func(t *testing.T) {
		t.Run("between identical trees", func(t *testing.T) {
			matches, err := SelectAll("div", []*dom.Node{doc, doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
		})
		t.Run("between a superset and subset", func(t *testing.T) {
			matches, err := SelectAll("p", []*dom.Node{doc, doc.Children[0]}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
		})
		t.Run("betweeen a subset and superset", func(t *testing.T) {
			matches, err := SelectAll("p", []*dom.Node{doc.Children[0], doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
		})
	})

	t.Run("selectAll", func(t *testing.T) {
		t.Run("should query array elements directly when they have no parents", func(t *testing.T) {
			divs := []*dom.Node{doc}
			matches, err := SelectAll("div", divs, nil)
			assert.Nil(t, err)
			assert.Equal(t, divs, matches)
		})
		t.Run("should query array elements directly when they have parents", func(t *testing.T) {
			ps, err := SelectAll("p", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			ps2, err := SelectAll("p", ps, nil)
			assert.Nil(t, err)
			assert.Equal(t, ps, ps2)
		})
		t.Run("should support pseudos led by a traversal (#111)", func(t *testing.T) {
			dom2 := parseDOM(`<div><div class="foo">a</div><div class="bar">b</div></div>`, false)[0]
			a, err := SelectAll(".foo:has(+.bar)", []*dom.Node{dom2}, nil)
			assert.Nil(t, err)
			assert.Len(t, a, 1)
			assert.Equal(t, dom2.Children[0], a[0])
		})
		t.Run("should accept document root nodes", func(t *testing.T) {
			doc := parseDOM("<div id=foo><p>foo</p></div>", false)
			matches, err := SelectAll(":contains(foo)", doc, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
		})
		t.Run("should support scoped selections relative to the root (#709)", func(t *testing.T) {
			doc := parseDOM(`
                <div class="parent">
                    <div class="one"><p class="p1"></p></div>
                    <div class="two"><p class="p2"></p></div>
                    <div class="three"><p class="p3"></p></div>
                </div>`, false)

			two, err := SelectOne(".two", doc, nil)
			assert.Nil(t, err)
			three, err := SelectOne(".parent .two .p2", []*dom.Node{two}, &types.Options{RelativeSelector: types.OptNo})
			assert.Nil(t, err)
			copy, err := three.CloneNode(false)
			assert.Nil(t, err)
			assert.Equal(t, dom.NewElement("p", orderedmap.NewOrderedMapWithElements(&orderedmap.Element[string, string]{Key: "class", Value: "p2"}), nil, dom.ElementTypeTag), copy)

			four, err := SelectOne(".parent .two .p3", []*dom.Node{two}, &types.Options{RelativeSelector: types.OptNo})
			assert.Nil(t, err)
			assert.Nil(t, four)
		})
		t.Run("cannot query element within template context, but still query template itself", func(t *testing.T) {
			doc := parseDOM(`<template><div><p id="insert"></p></div></template>`, false)

			matches, err := SelectAll("#insert", doc, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 0)
			one, err := SelectOne("#insert", doc, nil)
			assert.Nil(t, err)
			assert.Nil(t, one)
			matches, err = SelectAll("template", doc, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			one, err = SelectOne("template", doc, nil)
			assert.Nil(t, err)
			assert.NotNil(t, one)

			opts := &types.Options{XmlMode: types.OptYes}
			matches, err = SelectAll("#insert", doc, opts)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			one, err = SelectOne("#insert", doc, opts)
			assert.Nil(t, err)
			assert.NotNil(t, one)
			matches, err = SelectAll("template", doc, opts)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			one, err = SelectOne("template", doc, opts)
			assert.Nil(t, err)
			assert.NotNil(t, one)
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("should throw with a pseudo-element", func(t *testing.T) {
			_, err := Compile("::after", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "not supported")
		})

		t.Run("should throw an error if encountering a traversal-first selector with relative selectors disabled", func(t *testing.T) {
			_, err := Compile("> p", &types.Options{RelativeSelector: types.OptNo}, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "relative selectors are not allowed when the `relativeSelector` option is disabled")
		})

		t.Run("should throw with a column combinator", func(t *testing.T) {
			_, err := Compile("foo || bar", &types.Options{RelativeSelector: types.OptNo}, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
		})

		t.Run("should throw with attribute namespace", func(t *testing.T) {
			_, err := Compile("[foo|bar]", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
			//_, err = Compile("[|bar]", nil, nil)
			//assert.NotNil(t, err)
			//assert.Contains(t, err.Error(), notYet)
			_, err = Compile("[*|bar]", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
		})

		t.Run("should throw with tag namespace", func(t *testing.T) {
			_, err := Compile("foo|bar", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
			_, err = Compile("|bar", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
			_, err = Compile("*|bar", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
		})

		t.Run("should throw with universal selector", func(t *testing.T) {
			_, err := Compile("foo|*", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
			_, err = Compile("|*", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), notYet)
			//_, err = Compile("*|*", nil, nil)
			//assert.NotNil(t, err)
			//assert.Contains(t, err.Error(), notYet)
		})

		t.Run("should throw if parameter is supplied for pseudo", func(t *testing.T) {
			_, err := Compile(":any-link(test)", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "doesn't have any arguments")

			_, err = Compile(":only-child(test)", nil, nil)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "doesn't have any arguments")
		})

		t.Run("should throw if no parameter is supplied for pseudo", func(t *testing.T) {
			options := &types.Options{
				Pseudos: map[string]func(elem *dom.Node, value string) bool{
					"foovalue": func(elem *dom.Node, subselect string) bool {
						return domutils.GetAttributeValue(elem, "foo") == subselect
					},
				},
			}

			_, err := Compile(":foovalue", options, nil)
			assert.Nil(t, err) // we can't change the number of arguments in function
			// assert.Contains(t, err.Error(), "requires an argument")
		})
	})

	t.Run("unsatisfiable and universally valid selectors", func(t *testing.T) {
		t.Run("in :not", func(t *testing.T) {
			query, err := compileUnsafe(":not(*)", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
			query, err = compileUnsafe(":not(:not(:not(*)))", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
		})
		t.Run("in :has", func(t *testing.T) {
			matches, err := SelectAll(":has(*)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, doc, matches[0])

			matches2, err := SelectAll("p:has(+ *)", parseDOM("<p><p>", false), nil)
			assert.Nil(t, err)
			assert.Len(t, matches2, 1)
			assert.Equal(t, "p", matches2[0].TagName())
		})
		t.Run("in :is", func(t *testing.T) {
			query, err := compileUnsafe(":is(*)", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysTrue, query.Type)
			query, err = compileUnsafe(":is(:not(:not(*)))", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysTrue, query.Type)
			query, err = compileUnsafe(":is(*, :scope)", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysTrue, query.Type)
		})

		t.Run("should skip unsatisfiable", func(t *testing.T) {
			query, err := compileUnsafe("* :not(*) foo", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysFalse, query.Type)
		})

		t.Run("should promote universally valid", func(t *testing.T) {
			query, err := compileUnsafe("*, foo", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysTrue, query.Type)
		})

		t.Run("should promote `rootFunc`", func(t *testing.T) {
			query, err := compileUnsafe(":is(*), foo", nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, types.MatchTypeAlwaysTrue, query.Type)
		})
	})

	t.Run(":matches", func(t *testing.T) {
		t.Run("should select multiple elements", func(t *testing.T) {
			matches, err := SelectAll(":matches(p, div)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":matches(div, :not(div))", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":matches(boo, baa, tag, div, foo, bar, baz)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, doc, matches[0])
		})

		t.Run("should support traversals", func(t *testing.T) {
			matches, err := SelectAll(":matches(div p)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "p", matches[0].Name)

			matches, err = SelectAll(":matches(div > p)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "p", matches[0].Name)

			matches, err = SelectAll(":matches(p < div)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "div", matches[0].Name)

			matches, err = SelectAll(":matches(> p)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "p", matches[0].Name)

			matches, err = SelectAll("div:has(:is(:scope p))", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "div", matches[0].Name)

			multiLevelDom := parseDOM("<a><b><c><d>", false)[0]
			matches, err = SelectAll(":is(* c)", []*dom.Node{multiLevelDom}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, "c", matches[0].Name)
		})

		t.Run("should support alias :is", func(t *testing.T) {
			matches, err := SelectAll(":is(p, div)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":is(div, :not(div))", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":is(boo, baa, tag, div, foo, bar, baz)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, doc, matches[0])
		})

		t.Run("should support alias :where", func(t *testing.T) {
			matches, err := SelectAll(":where(p, div)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":where(div, :not(div))", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 2)
			matches, err = SelectAll(":where(boo, baa, tag, div, foo, bar, baz)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, doc, matches[0])
		})
	})

	t.Run("parent selector (<)", func(t *testing.T) {
		t.Run("should select the right element", func(t *testing.T) {
			matches, err := SelectAll("p < div", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
			assert.Equal(t, doc, matches[0])
		})
		t.Run("should not select nodes without children", func(t *testing.T) {
			matches, err := SelectAll("p < div", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			matches2, err := SelectAll("* < *", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Equal(t, matches2, matches)
		})
	})

	t.Run("selectOne", func(t *testing.T) {
		t.Run("should select elements in traversal order", func(t *testing.T) {
			one, err := SelectOne("p", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Equal(t, doc.Children[0], one)
			one, err = SelectOne(":contains(foo)", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Equal(t, doc, one)
		})
		t.Run("should take shortcuts when applicable", func(t *testing.T) {
			match, err := SelectOne("*", []*dom.Node{}, nil)
			assert.Nil(t, err)
			assert.Nil(t, match)
		})
		t.Run("should properly handle root elements", func(t *testing.T) {
			match, err := SelectOne("div:root", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Equal(t, doc, match)
			match, err = SelectOne("* > div", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Nil(t, match)
		})
	})

	t.Run("options", func(t *testing.T) {
		t.Run("should recognize contexts", func(t *testing.T) {
			div, err := SelectAll("div", []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			p, err := SelectAll("p", []*dom.Node{doc}, nil)
			assert.Nil(t, err)

			match, err := SelectOne("div", div, &types.Options{Context: div})
			assert.Nil(t, err)
			assert.Equal(t, div[0], match)
			match, err = SelectOne("div", div, &types.Options{Context: p})
			assert.Nil(t, err)
			assert.Nil(t, match)
			match2, err := SelectAll("p", div, &types.Options{Context: div})
			assert.Equal(t, p, match2)
		})

		t.Run("should not crash when siblings repeat", func(t *testing.T) {
			dom := parseDOM(strings.Repeat(`<div></div>`, 51), false)

			matches, err := SelectAll("+div", dom, &types.Options{Context: dom})
			assert.Nil(t, err)
			assert.Len(t, matches, 50)
		})

		t.Run("should cache results by default", func(t *testing.T) {
			doc := parseDOM(`<div><p id="foo">bar</p></div>`, false)[0]
			selector := ":has(#bar) p"

			matches, err := SelectAll(selector, []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 0)
			doc.Children[0].Attribs.Set("id", "bar")

			matches, err = SelectAll(selector, []*dom.Node{doc}, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
		})
	})

	t.Run("optional adapter methods", func(t *testing.T) {
		t.Run("should support prevElementSibling", func(t *testing.T) {
			dom := parseDOM(strings.Repeat("<p>foo", 10)+`<div>bar</div>`, false)

			matches, err := SelectAll("p + div", dom, nil)
			assert.Nil(t, err)
			assert.Len(t, matches, 1)
		})
	})
}
