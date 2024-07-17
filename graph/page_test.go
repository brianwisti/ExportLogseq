package graph_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
)

func TestPage_NewEmptyPage(t *testing.T) {
	page := graph.NewEmptyPage()

	assert.NotNil(t, page)
	assert.Equal(t, "page", page.Kind)
	assert.NotNil(t, page.Root)
	assert.Contains(t, page.AllBlocks, page.Root)
}

func TestPage_Aliases_Empty(t *testing.T) {
	page := graph.NewEmptyPage()
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
		page := graph.NewEmptyPage()
		page.Root.Properties.Set("alias", tt.PropValue)
		aliases := page.Aliases()

		assert.ElementsMatch(t, tt.want, aliases)
	}
}

func TestPage_IsJournal(t *testing.T) {
	page := graph.NewEmptyPage()
	page.Name = "2022-01-01"
	page.Title = page.Name

	assert.True(t, page.IsJournal())
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
		page := graph.NewEmptyPage()
		page.PathInGraph = tt.PathInGraph

		assert.Equal(t, tt.want, page.IsPlaceholder())
	}
}

func TestPage_IsPublic_Default(t *testing.T) {
	page := graph.NewEmptyPage()

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

	page := graph.NewEmptyPage()
	for _, tt := range isPublicTests {
		page.Root.Properties.Set("public", tt.PropValue)

		assert.Equal(t, tt.want, page.IsPublic())
	}

	assert.False(t, page.IsPublic())
}

func TestPage_Links_Empty(t *testing.T) {
	page := graph.NewEmptyPage()

	assert.Empty(t, page.Links())
}

func TestPage_Links_FromRoot(t *testing.T) {
	page := graph.NewEmptyPage()
	link := graph.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: graph.LinkTypeResource,
		IsEmbed:  false,
	}
	link, _ = page.Root.Content.AddLink(link)
	links := page.Links()

	assert.Contains(t, links, link)
}

func TestPage_Properties_Empty(t *testing.T) {
	page := graph.NewEmptyPage()
	pageProps := page.Properties()

	assert.NotNil(t, pageProps)
	assert.Empty(t, pageProps.Properties)
}

func TestPage_Properties_FromRoot(t *testing.T) {
	page := graph.NewEmptyPage()
	page.Name = "Test Page"
	page.Root.Properties.Set("id", "123")
	pageProps := page.Properties()
	got, ok := pageProps.Get("id")

	assert.True(t, ok)
	assert.Equal(t, "123", got.Value)
}

func TestPage_SetRoot(t *testing.T) {
	page := graph.NewEmptyPage()
	oldRoot := page.Root
	root := graph.NewEmptyBlock()
	assert.NotEqual(t, root, oldRoot)
	page.SetRoot(root)

	assert.Equal(t, root, page.Root)
	assert.NotContains(t, page.AllBlocks, oldRoot)
	assert.Contains(t, page.AllBlocks, page.Root)
}

func TestPage_SetRoot_WithChildren(t *testing.T) {
	page := graph.NewEmptyPage()
	root := graph.NewEmptyBlock()
	child := graph.NewEmptyBlock()
	root.Children = []*graph.Block{child}
	page.SetRoot(root)

	assert.Equal(t, root, page.Root)
	assert.Contains(t, page.AllBlocks, root)
	assert.Contains(t, page.AllBlocks, child)
}
