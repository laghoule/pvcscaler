package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func (t *testing.T) {
	tests := []struct {
		name       string
		namespaces []string
		expected   bool
	}{
		{
			name:       "valid namespace",
			namespaces: []string{"default"},
			expected:   true,
		},
		{
			name:       "invalid namespace",
			namespaces: []string{"all", "default"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, validNamespaces(tt.namespaces))
		})
	}
}
