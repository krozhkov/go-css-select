package helpers

import "github.com/krozhkov/go-css-select/query/types"

/**
 * Create a copy of options, omitting `context` and `rootFunc`.
 *
 * This is used when compiling nested selectors (e.g. inside `:is`, `:not`,
 * `:nth-child(… of S)`) so that the parent compilation state doesn't leak.
 */
func CopyOptions(
	options *types.Options,
) *types.Options {
	if options == nil {
		return &types.Options{}
	}

	opts := *options
	// Omit context and rootFunc so parent compilation state doesn't leak.
	opts.Context = nil
	opts.RootFunc = nil

	return &opts
}
