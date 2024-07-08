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

func TestGraph_FindPage(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage("Test Page")

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_FindPage_NotFound(t *testing.T) {
	graph := logseq.NewGraph()
	_, err := graph.FindPage("Test Page")

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.PageNotFoundError{PageName: "Test Page"})
}

func TestGraph_FindPage_WithAlias(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	page.Root.Properties.Set("alias", "alias")
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage("alias")

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}
