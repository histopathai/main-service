package cache

import (
	"strings"
)

type KeyBuilder struct {
	prefix string
}

func NewKeyBuilder(prefix string) *KeyBuilder {
	return &KeyBuilder{prefix: prefix}
}

func (kb *KeyBuilder) Build(parts ...string) string {
	allParts := make([]string, 0, len(parts)+1)

	if kb.prefix != "" {
		allParts = append(allParts, kb.prefix)
	}

	allParts = append(allParts, parts...)

	return strings.Join(allParts, ":")
}

func (kb *KeyBuilder) BuildPattern(parts ...string) string {
	return kb.Build(parts...)
}
