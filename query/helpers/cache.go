package helpers

import (
	"runtime"
	"sync"
	"weak"

	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
)

type Cache[K any, V any] struct {
	store map[weak.Pointer[K]]V
	mu    sync.Mutex
}

func NewCache[K any, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		store: make(map[weak.Pointer[K]]V),
	}
}

func (c *Cache[K, V]) Set(key *K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ptr := weak.Make(key)

	c.store[ptr] = value

	runtime.AddCleanup(key, func(p weak.Pointer[K]) {
		c.mu.Lock()
		delete(c.store, p)
		c.mu.Unlock()
	}, ptr)
}

func (c *Cache[K, V]) Has(key *K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	ptr := weak.Make(key)

	_, ok := c.store[ptr]

	return ok
}

func (c *Cache[K, V]) Get(key *K) V {
	c.mu.Lock()
	defer c.mu.Unlock()

	ptr := weak.Make(key)

	value, _ := c.store[ptr]

	return value
}

func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.store)
}

/**
 * Some selectors such as `:contains` and (non-relative) `:has` will only be
 * able to match elements if their parents match the selector (as they contain
 * a subset of the elements that the parent contains).
 *
 * This function wraps the given `matches` function in a function that caches
 * the results of the parent elements, so that the `matches` function only
 * needs to be called once for each subtree.
 */
func CacheParentResults(
	next *types.CompiledQuery,
	options *types.Options,
	matches func(elem *dom.Node) bool,
) *types.CompiledQuery {
	if options != nil && options.CacheResults == types.OptNo {
		return &types.CompiledQuery{
			Match: func(elem *dom.Node) bool {
				return next.Match(elem) && matches(elem)
			},
		}
	}

	// Use a cache to avoid re-checking children of an element.
	resultCache := NewCache[dom.Node, bool]()

	addResultToCache := func(elem *dom.Node) bool {
		result := matches(elem)

		resultCache.Set(elem, result)
		return result
	}

	return &types.CompiledQuery{
		Match: func(elem *dom.Node) bool {
			if !next.Match(elem) {
				return false
			}
			if resultCache.Has(elem) {
				return resultCache.Get(elem)
			}

			// Check all of the element's parents.
			node := elem

			for {
				parent := GetElementParent(node)

				if parent == nil {
					return addResultToCache(elem)
				}

				node = parent

				if resultCache.Has(node) {
					break
				}
			}

			return resultCache.Get(node) && addResultToCache(elem)
		},
	}
}
