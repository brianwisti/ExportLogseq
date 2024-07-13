package graph_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
)

func TestNewEmptyBlockContent(t *testing.T) {
	content := graph.NewEmptyBlockContent()

	assert.NotNil(t, content)
	assert.Empty(t, content.BlockID)
	assert.Empty(t, content.Links)
}

func TestNewBlockContent(t *testing.T) {
	block := graph.NewEmptyBlock()
	rawSource := ""
	content, err := graph.NewBlockContent(block, rawSource)

	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, block.ID, content.BlockID)
}

func TestBlockContent_AddLink(t *testing.T) {
	addLinkTests := []struct {
		LinkType graph.LinkType
		linkPath string
		Label    string
		IsEmbed  bool
	}{
		{graph.LinkTypeResource, gofakeit.URL(), gofakeit.Phrase(), false},
	}

	for _, tt := range addLinkTests {
		content := graph.NewEmptyBlockContent()
		link := graph.Link{
			LinkPath: tt.linkPath,
			Label:    tt.Label,
			LinkType: tt.LinkType,
			IsEmbed:  tt.IsEmbed,
		}

		addedLink, err := content.AddLink(link)

		assert.NoError(t, err)
		assert.NotEmpty(t, content.Links)

		assert.Equal(t, content.BlockID, addedLink.LinksFrom)
		assert.Equal(t, link.LinkPath, addedLink.LinkPath)
		assert.Equal(t, link.Label, addedLink.Label)
		assert.Equal(t, link.LinkType, addedLink.LinkType)
		assert.Equal(t, link.IsEmbed, addedLink.IsEmbed)
	}
}

func TestBlockContent_AddLink_Duplicate(t *testing.T) {
	content := graph.NewEmptyBlockContent()
	link := graph.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: graph.LinkTypeResource,
		IsEmbed:  false,
	}
	content.AddLink(link)
	linkCount := len(content.Links)
	addedLink, err := content.AddLink(link)

	assert.Empty(t, addedLink)
	assert.NoError(t, err)
	assert.Equal(t, linkCount, len(content.Links))
}

func TestBlockContent_FindLink(t *testing.T) {
	content := graph.NewEmptyBlockContent()
	linkPath := gofakeit.URL()
	link := graph.Link{
		LinksFrom: content.BlockID,
		LinkPath:  linkPath,
		Label:     gofakeit.Phrase(),
		LinkType:  graph.LinkTypeResource,
		IsEmbed:   false,
	}
	content.Links[linkPath] = link
	foundLink, ok := content.FindLink(linkPath)

	assert.True(t, ok)
	assert.NotNil(t, foundLink)
	assert.Equal(t, linkPath, foundLink.LinkPath)
}

func TestBlockContent_FindLink_NotFound(t *testing.T) {
	content := graph.NewEmptyBlockContent()
	linkPath := gofakeit.URL()
	foundLink, ok := content.FindLink(linkPath)

	assert.False(t, ok)
	assert.Empty(t, foundLink)
}

func TestBlockContent_IsCodeBlock(t *testing.T) {
	codeBlockMarkdown := "```\ncode\n```"
	content := graph.BlockContent{
		Markdown: codeBlockMarkdown,
	}
	assert.True(t, content.IsCodeBlock())
}

func TestBlockContent_SetMarkdown(t *testing.T) {
	content := graph.NewEmptyBlockContent()
	markdown := gofakeit.Sentence(5)
	err := content.SetMarkdown(markdown)

	assert.NoError(t, err)
	assert.Equal(t, markdown, content.Markdown)
}
