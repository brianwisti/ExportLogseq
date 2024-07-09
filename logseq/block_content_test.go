package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestNewEmptyBlockContent(t *testing.T) {
	content := logseq.NewEmptyBlockContent()

	assert.NotNil(t, content)
	assert.Empty(t, content.Block)
	assert.Empty(t, content.ResourceLinks)
}

func TestNewBlockContent(t *testing.T) {
	block := logseq.NewEmptyBlock()
	rawSource := ""
	content := logseq.NewBlockContent(block, rawSource)

	assert.NotNil(t, content)
	assert.Equal(t, block, content.Block)
}

func TestBlockContent_AddLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := logseq.ExternalResource{Uri: "https://example.com"}
	link, err := content.AddLinkToResource(resource)

	assert.NoError(t, err)
	assert.NotEmpty(t, content.ResourceLinks)
	assert.Equal(t, resource, link.LinksTo)
	assert.False(t, link.IsEmbed)
}

func TestBlockContent_AddLinkToResource_Duplicate(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := logseq.ExternalResource{Uri: "https://example.com"}
	content.AddLinkToResource(resource)
	resourceCount := len(content.ResourceLinks)
	link, err := content.AddLinkToResource(resource)

	assert.Nil(t, link)
	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.ErrorDuplicateResourceLink{Resource: resource})
	assert.Equal(t, resourceCount, len(content.ResourceLinks))
}

func TestBlockContent_AddEmbeddedLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := logseq.ExternalResource{Uri: "https://example.com"}
	link, err := content.AddEmbeddedLinkToResource(resource)

	assert.NoError(t, err)
	assert.NotEmpty(t, content.ResourceLinks)
	assert.Equal(t, resource, link.LinksTo)
	assert.True(t, link.IsEmbed)
}

func TestBlockContent_FindLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := logseq.ExternalResource{Uri: "https://example.com"}
	content.AddLinkToResource(resource)
	link := content.FindLinkToResource(resource)

	assert.Equal(t, resource, link.LinksTo)
}

func TestBlockContent_FindLinkToResource_NotFound(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := logseq.ExternalResource{Uri: "https://example.com"}

	assert.Nil(t, content.FindLinkToResource(resource))
}

func TestBlockContent_IsCodeBlock(t *testing.T) {
	codeBlockMarkdown := "```\ncode\n```"
	content := logseq.BlockContent{
		Markdown: codeBlockMarkdown,
	}
	assert.True(t, content.IsCodeBlock())
}
