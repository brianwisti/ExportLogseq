package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestPage_NewEmptyPage(t *testing.T) {
	page := logseq.NewEmptyPage()

	assert.NotNil(t, page)
	assert.Equal(t, "page", page.Kind)
	assert.NotNil(t, page.Root)
	assert.Contains(t, page.AllBlocks, page.Root)
}

func TestPage_InContext(t *testing.T) {
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	got, err := page.InContext(*logseq.NewGraph())

	assert.NoError(t, err)
	assert.Equal(t, "/test-page", got)
}
