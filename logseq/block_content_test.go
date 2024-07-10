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

func TestBlockContent_AddLinkToAsset(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	assetPath := "assets/test.jpg"
	label := LinkLabel()
	link, err := content.AddLinkToAsset(assetPath, label)

	assert.NoError(t, err)
	assert.NotNil(t, link)
}

func TestBlockContent_AddLinkToAsset_Duplicate(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	assetPath := "assets/test.jpg"
	label := LinkLabel()
	content.AddLinkToAsset(assetPath, label)
	assetCount := len(content.ResourceLinks)
	link, err := content.AddLinkToAsset(assetPath, label)

	assert.Nil(t, link)
	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.ErrorDuplicateAssetLink{AssetPath: assetPath})
	assert.Equal(t, assetCount, len(content.ResourceLinks))
}

func TestBlockContent_AddEmbeddedLinkToAsset(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	assetPath := "assets/test.jpg"
	label := LinkLabel()
	link, err := content.AddEmbeddedLinkToAsset(assetPath, label)

	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.True(t, link.IsEmbed)
}

func TestBlockContent_AddLinkToPage(t *testing.T) {
	fromPage := Page()
	content := fromPage.Root.Content
	pageName, label := PageName(), LinkLabel()
	link, err := content.AddLinkToPage(pageName, label)

	assert.NoError(t, err)
	assert.NotEmpty(t, link.LinksFrom)
	assert.NotEmpty(t, link.LinksTo)
	assert.NotEmpty(t, content.PageLinks)
	assert.Equal(t, label, link.Label)
	assert.False(t, link.IsEmbed)

	linkedPage := link.LinksTo.(*logseq.Page)
	assert.Equal(t, pageName, linkedPage.Name)
}

func TestBlockContent_AddLinkToPage_Duplicate(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	pageName, label := PageName(), LinkLabel()
	content.AddLinkToPage(pageName, label)
	pageCount := len(content.PageLinks)
	link, err := content.AddLinkToPage(pageName, label)

	assert.Nil(t, link)
	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.ErrorDuplicatePageLink{PageName: pageName})
	assert.Equal(t, pageCount, len(content.PageLinks))
}

func TestBlockContent_AddLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource, label := ExternalResource(), LinkLabel()
	link, err := content.AddLinkToResource(resource, label)

	assert.NoError(t, err)
	assert.NotEmpty(t, content.ResourceLinks)
	assert.Equal(t, resource, link.LinksTo)
	assert.Equal(t, label, link.Label)
	assert.False(t, link.IsEmbed)
}

func TestBlockContent_AddLinkToResource_Duplicate(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource, label := ExternalResource(), LinkLabel()
	content.AddLinkToResource(resource, label)
	resourceCount := len(content.ResourceLinks)
	link, err := content.AddLinkToResource(resource, label)

	assert.Nil(t, link)
	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.ErrorDuplicateResourceLink{Resource: resource})
	assert.Equal(t, resourceCount, len(content.ResourceLinks))
}

func TestBlockContent_AddEmbeddedLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource, label := ExternalResource(), LinkLabel()
	link, err := content.AddEmbeddedLinkToResource(resource, label)

	assert.NoError(t, err)
	assert.NotEmpty(t, content.ResourceLinks)
	assert.Equal(t, resource, link.LinksTo)
	assert.True(t, link.IsEmbed)
}

func TestBlockContent_FindLinkToAsset(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	assetPath := "assets/test.jpg"
	label := LinkLabel()
	content.AddLinkToAsset(assetPath, label)
	link := content.FindLinkToAsset(assetPath)

	assert.NotNil(t, link)
	assert.Equal(t, assetPath, link.LinksTo.(*logseq.Asset).PathInGraph)
}

func TestBlockContent_FindLinkToPage(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	pageName, label := PageName(), LinkLabel()
	content.AddLinkToPage(pageName, label)
	link := content.FindLinkToPage(pageName)

	assert.NotNil(t, link)
	assert.Equal(t, pageName, link.LinksTo.(*logseq.Page).Name)
}

func TestBlockContent_FindLinkToResource(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource, label := ExternalResource(), LinkLabel()
	content.AddLinkToResource(resource, label)
	link := content.FindLinkToResource(resource)

	assert.Equal(t, resource, link.LinksTo)
}

func TestBlockContent_FindLinkToResource_NotFound(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	resource := ExternalResource()

	assert.Nil(t, content.FindLinkToResource(resource))
}

func TestBlockContent_IsCodeBlock(t *testing.T) {
	codeBlockMarkdown := "```\ncode\n```"
	content := logseq.BlockContent{
		Markdown: codeBlockMarkdown,
	}
	assert.True(t, content.IsCodeBlock())
}

func TestBlockContent_SetMarkdown(t *testing.T) {
	content := logseq.NewEmptyBlockContent()
	markdown := "test"
	err := content.SetMarkdown(markdown)

	assert.NoError(t, err)
	assert.Equal(t, markdown, content.Markdown)
}
