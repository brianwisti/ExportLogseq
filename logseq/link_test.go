package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestExternalResource_InContext(t *testing.T) {
	graph := logseq.NewGraph()
	resource := ExternalResource()
	path, err := resource.InContext(*graph)

	assert.NoError(t, err)
	assert.Equal(t, resource.Uri, path)
}
