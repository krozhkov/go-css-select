package pseudoselectors

import (
	"fmt"

	"github.com/krozhkov/go-css-select/parser"
	"github.com/krozhkov/go-css-select/query/types"
	"github.com/krozhkov/go-htmlparser2/dom"
)

/**
 * Compile a pseudo selector into an executable query function.
 * @param next Matcher to run after this matcher succeeds.
 * @param selector Selector used to match elements.
 * @param options Options that control this operation.
 * @param context Context nodes used to scope selector matching.
 * @param compileToken Function used to compile nested selector tokens.
 */
func CompilePseudoSelector(
	next *types.CompiledQuery,
	selector *parser.Selector,
	options *types.Options,
	context []*dom.Node,
	compileToken types.CompileToken,
) (*types.CompiledQuery, error) {
	name := selector.Name
	children := selector.Children
	var data string
	if selector.Data != nil {
		data = *selector.Data
	}

	if len(children) > 0 {
		subselect, ok := subselects[name]
		if !ok {
			return nil, fmt.Errorf("unknown pseudo-class :%s", name)
		}

		return subselect(next, children, options, context, compileToken)
	}

	var userPseudos map[string]func(elem *dom.Node, value string) bool
	if options != nil && options.Pseudos != nil {
		userPseudos = options.Pseudos
	}

	if userPseudo, ok := userPseudos[name]; ok {
		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				return userPseudo(element, data) && next.Match(element, scope)
			},
		}, nil
	}

	if stringPseudo, ok := aliases[name]; ok {
		if data != "" {
			return nil, fmt.Errorf("pseudo %s doesn't have any arguments", name)
		}

		// The alias has to be parsed here, to make sure options are respected.
		alias, err := parser.Parse(stringPseudo)
		if err != nil {
			return nil, err
		}
		return is(next, alias, options, context, compileToken)
	}

	if filterPseudo, ok := filters[name]; ok {
		return filterPseudo(next, data, options, context, compileToken)
	}

	if pseudo, ok := pseudos[name]; ok {
		if data != "" {
			return nil, fmt.Errorf("pseudo-class :%s doesn't have any arguments", name)
		}

		return &types.CompiledQuery{
			Match: func(element *dom.Node, scope *dom.Node) bool {
				return pseudo(element, options) && next.Match(element, scope)
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown pseudo-class :%s", name)
}
