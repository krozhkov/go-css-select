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
	mu    sync.RWMutex
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
	_, exists := c.store[ptr]
	c.store[ptr] = value

	if !exists {
		runtime.AddCleanup(key, func(p weak.Pointer[K]) {
			c.mu.Lock()
			delete(c.store, p)
			c.mu.Unlock()
		}, ptr)
	}
}

func (c *Cache[K, V]) Get(key *K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ptr := weak.Make(key)

	value, ok := c.store[ptr]

	return value, ok
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
 * @param next Matcher to run after this matcher succeeds.
 * @param options Configuration object for cache behavior.
 * @param options.cacheResults Whether results should be memoized by input root.
 * @param matches Compiled matcher function to wrap with caching.
 */
func CacheParentResults(
	next *types.CompiledQuery,
	options *types.Options,
	matches func(element *dom.Node, scope *dom.Node) bool,
) *types.CompiledQuery {
	if options != nil && options.CacheResults == types.OptNo {
		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				return next.Match(element, scope) && matches(element, scope)
			},
		}
	}

	// Use a cache to avoid re-checking children of an element.
	resultCache := NewCache[dom.Node, bool]()

	addResultToCache := func(element *dom.Node, scope *dom.Node) bool {
		result := matches(element, scope)

		resultCache.Set(element, result)
		return result
	}

	return &types.CompiledQuery{
		Match: func(element *dom.Node, scope *dom.Node) bool {
			if !next.Match(element, scope) {
				return false
			}
			if cached, ok := resultCache.Get(element); ok {
				return cached
			}

			// Check all of the element's parents.
			node := element
			var result bool
			var found bool

			for {
				parent := GetElementParent(node)

				if parent == nil {
					return addResultToCache(element, scope)
				}

				node = parent

				if result, found = resultCache.Get(node); found {
					break
				}
			}

			if result {
				return addResultToCache(element, scope)
			}

			return false
		},
	}
}
