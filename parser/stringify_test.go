package parser

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringifyOwnTests(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			serialized := Stringify(tt.expected)
			parsed, err := Parse(serialized)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func TestStringifyCollectedSelectors(t *testing.T) {
	var out map[string][][]*Selector

	err := json.Unmarshal(testData, &out)
	assert.Nil(t, err)

	for selector, expected := range out {
		t.Run(selector, func(t *testing.T) {
			serialized := Stringify(expected)
			parsed, err := Parse(serialized)
			assert.Nil(t, err)
			assert.Equal(t, expected, parsed)
		})
	}
}
