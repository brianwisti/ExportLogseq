package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestGraph_NewGraph(t *testing.T) {
	graph := logseq.NewGraph()

	assert.NotNil(t, graph)
}

func TestGraph_AddPage(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	err := graph.AddPage(&page)
	assert.NoError(t, err)
	addedPage, ok := graph.Pages[page.Name]

	assert.True(t, ok)
	assert.Equal(t, &page, addedPage)
}

func TestGraph_AddPage_WithExistingPage(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	_ = graph.AddPage(&page)
	err := graph.AddPage(&page)

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.PageExistsError{PageName: page.Name})
}
