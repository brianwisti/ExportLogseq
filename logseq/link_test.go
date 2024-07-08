package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestExternalResource_InContext(t *testing.T) {
	uri := "https://example.com"
	resource := logseq.ExternalResource{
		Uri: uri,
	}

	g := logseq.Graph{
		Pages: map[string]*logseq.Page{},
	}

	path, err := resource.InContext(g)
	assert.NoError(t, err)
	assert.Equal(t, uri, path)
}
