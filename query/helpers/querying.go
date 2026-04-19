package helpers

import (
	"slices"

	"github.com/krozhkov/go-css-select/query/internal"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/domutils"
)

/**
 * Find all elements matching the query. If not in XML mode, the query will ignore
 * the contents of `<template>` elements.
 *
 * @param query - Function that returns true if the element matches the query.
 * @param elems - Nodes to query. If a node is an element, its children will be queried.
 * @param options - Options for querying the document.
 * @returns All matching elements.
 */
func FindAll(
	query func(*dom.Node) bool,
	elems []*dom.Node,
	options *types.Options,
) []*dom.Node {
	var xmlMode bool
	if options != nil && options.XmlMode == types.OptYes {
		xmlMode = true
	}

	result := []*dom.Node{}
	/** Stack of the arrays we are looking at. */
	nodeStack := [][]*dom.Node{elems}
	/** Stack of the indices within the arrays. */
	indexStack := []int{0}

	for {
		length := len(indexStack)
		// First, check if the current array has any more elements to look at.
		if indexStack[length-1] >= len(nodeStack[length-1]) {
			// If we have no more arrays to look at, we are done.
			if len(indexStack) == 1 {
				return result
			}

			nodeStack[length-1] = nil
			nodeStack = nodeStack[:length-1]
			indexStack = indexStack[:length-1]

			// Loop back to the start to continue with the next array.
			continue
		}

		length = len(indexStack)
		elem := nodeStack[length-1][indexStack[length-1]]
		indexStack[length-1]++

		if !dom.IsTag(elem) {
			continue
		}
		if query(elem) {
			result = append(result, elem)
		}

		if xmlMode || domutils.GetName(elem) != "template" {
			/*
			 * Add the children to the stack. We are depth-first, so this is
			 * the next array we look at.
			 */
			children := domutils.GetChildren(elem)

			if len(children) > 0 {
				nodeStack = append(nodeStack, children)
				indexStack = append(indexStack, 0)
			}
		}
	}
}

/**
 * Find the first element matching the query. If not in XML mode, the query will ignore
 * the contents of `<template>` elements.
 *
 * @param query - Function that returns true if the element matches the query.
 * @param elems - Nodes to query. If a node is an element, its children will be queried.
 * @param options - Options for querying the document.
 * @returns The first matching element, or null if there was no match.
 */
func FindOne(
	query func(*dom.Node) bool,
	elems []*dom.Node,
	options *types.Options,
) *dom.Node {
	var xmlMode bool
	if options != nil && options.XmlMode == types.OptYes {
		xmlMode = true
	}

	/** Stack of the arrays we are looking at. */
	nodeStack := [][]*dom.Node{elems}
	/** Stack of the indices within the arrays. */
	indexStack := []int{0}

	for {
		length := len(indexStack)
		// First, check if the current array has any more elements to look at.
		if indexStack[length-1] >= len(nodeStack[length-1]) {
			// If we have no more arrays to look at, we are done.
			if len(indexStack) == 1 {
				return nil
			}

			nodeStack[length-1] = nil
			nodeStack = nodeStack[:length-1]
			indexStack = indexStack[:length-1]

			// Loop back to the start to continue with the next array.
			continue
		}

		length = len(indexStack)
		elem := nodeStack[length-1][indexStack[length-1]]
		indexStack[length-1]++

		if !dom.IsTag(elem) {
			continue
		}
		if query(elem) {
			return elem
		}

		if xmlMode || domutils.GetName(elem) != "template" {
			/*
			 * Add the children to the stack. We are depth-first, so this is
			 * the next array we look at.
			 */
			children := domutils.GetChildren(elem)

			if len(children) > 0 {
				nodeStack = append(nodeStack, children)
				indexStack = append(indexStack, 0)
			}
		}
	}
}

func GetNextSiblings(
	elem *dom.Node,
) []*dom.Node {
	siblings := domutils.GetSiblings(elem)
	if len(siblings) <= 1 {
		return nil
	}
	elemIndex := slices.Index(siblings, elem)
	if elemIndex < 0 || elemIndex == len(siblings)-1 {
		return nil
	}

	return internal.FilterFunc(siblings[elemIndex+1:], dom.IsTag)
}

func GetElementParent(
	node *dom.Node,
) *dom.Node {
	parent := domutils.GetParent(node)
	if parent != nil && dom.IsTag(parent) {
		return parent
	}
	return nil
}
