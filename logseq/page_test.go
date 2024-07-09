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

func TestPage_Aliases_Empty(t *testing.T) {
	page := logseq.NewEmptyPage()
	aliases := page.Aliases()

	assert.NotNil(t, aliases)
	assert.Empty(t, aliases)
}

func TestPage_Aliases_FromProperties(t *testing.T) {
	aliasesTests := []struct {
		PropValue string
		want      []string
	}{
		{"a", []string{"a"}},
		{"a, b", []string{"a", "b"}},
		{"a, b, c", []string{"a", "b", "c"}},
	}

	for _, tt := range aliasesTests {
		page := logseq.NewEmptyPage()
		page.Root.Properties.Set("alias", tt.PropValue)
		aliases := page.Aliases()

		assert.ElementsMatch(t, tt.want, aliases)
	}
}

func TestPage_InContext(t *testing.T) {
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	page.PathInGraph = "/test-page"

	graph := *logseq.NewGraph()
	graph.AddPage(&page)
	got, err := page.InContext(graph)

	assert.NoError(t, err)
	assert.Equal(t, "/test-page", got)
}

func TestPage_InContext_WithGraphError(t *testing.T) {
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	page.PathInGraph = "/test-page"

	graph := *logseq.NewGraph()
	_, err := page.InContext(graph)

	assert.ErrorIs(t, err, logseq.DisconnectedPageError{PageName: "Test Page"})
}

func TestPage_IsPlaceholder(t *testing.T) {
	isPlaceholderTests := []struct {
		PathInGraph string
		want        bool
	}{
		{"", true},
		{"test-page", false},
	}

	for _, tt := range isPlaceholderTests {
		page := logseq.NewEmptyPage()
		page.PathInGraph = tt.PathInGraph

		assert.Equal(t, tt.want, page.IsPlaceholder())
	}
}

func TestPage_IsPublic_Default(t *testing.T) {
	page := logseq.NewEmptyPage()

	assert.False(t, page.IsPublic())
}

func TestPage_IsPublic_FromRoot(t *testing.T) {
	isPublicTests := []struct {
		PropValue string
		want      bool
	}{
		{"true", true},
		{"false", false},
		{"", false},
	}

	page := logseq.NewEmptyPage()
	for _, tt := range isPublicTests {
		page.Root.Properties.Set("public", tt.PropValue)

		assert.Equal(t, tt.want, page.IsPublic())
	}

	assert.False(t, page.IsPublic())
}

func TestPage_Properties_Empty(t *testing.T) {
	page := logseq.NewEmptyPage()
	pageProps := page.Properties()

	assert.NotNil(t, pageProps)
	assert.Empty(t, pageProps.Properties)
}

func TestPage_Properties_FromRoot(t *testing.T) {
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	page.Root.Properties.Set("id", "123")
	pageProps := page.Properties()
	got, ok := pageProps.Get("id")

	assert.True(t, ok)
	assert.Equal(t, "123", got.Value)
}

func TestPage_ResourceLinks(t *testing.T) {
	page := logseq.NewEmptyPage()
	block := logseq.NewEmptyBlock()
	resource := logseq.ExternalResource{Uri: "https://example.com"}
	label := "Example"
	link, _ := block.Content.AddLinkToResource(resource, label)
	page.SetRoot(block)
	links := page.ResourceLinks()

	assert.Contains(t, links, link)
}

func TestPage_SetRoot(t *testing.T) {
	page := logseq.NewEmptyPage()
	oldRoot := page.Root
	root := logseq.NewEmptyBlock()
	assert.NotEqual(t, root, oldRoot)
	page.SetRoot(root)

	assert.Equal(t, root, page.Root)
	assert.NotContains(t, page.AllBlocks, oldRoot)
	assert.Contains(t, page.AllBlocks, page.Root)
}

func TestPage_SetRoot_WithChildren(t *testing.T) {
	page := logseq.NewEmptyPage()
	root := logseq.NewEmptyBlock()
	child := logseq.NewEmptyBlock()
	root.Children = []*logseq.Block{child}
	page.SetRoot(root)

	assert.Equal(t, root, page.Root)
	assert.Contains(t, page.AllBlocks, root)
	assert.Contains(t, page.AllBlocks, child)
}
