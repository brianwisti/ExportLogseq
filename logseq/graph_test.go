package logseq_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestGraph_NewGraph(t *testing.T) {
	graph := logseq.NewGraph()

	assert.NotNil(t, graph)
	assert.Empty(t, graph.Pages)
	assert.Empty(t, graph.Assets)
}

func TestGraph_AddAsset(t *testing.T) {
	graph := logseq.NewGraph()
	asset := logseq.NewAsset("assets/test.jpg")
	err := graph.AddAsset(&asset)

	assert.NoError(t, err)
	assetKey := strings.ToLower(asset.PathInGraph)
	addedAsset, ok := graph.Assets[assetKey]

	assert.True(t, ok)
	assert.Equal(t, &asset, addedAsset)
}

func TestGraph_AddAsset_WithExistingAsset(t *testing.T) {
	graph := logseq.NewGraph()
	asset := logseq.NewAsset("assets/test.jpg")
	_ = graph.AddAsset(&asset)
	err := graph.AddAsset(&asset)

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.AssetExistsError{AssetPath: asset.PathInGraph})
}

func TestGraph_AddPage(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	err := graph.AddPage(&page)
	assert.NoError(t, err)
	pageKey := strings.ToLower(page.Name)
	addedPage, ok := graph.Pages[pageKey]

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

func TestGraph_FindAsset(t *testing.T) {
	graph := logseq.NewGraph()
	asset := logseq.NewAsset("assets/test.jpg")
	_ = graph.AddAsset(&asset)
	foundAsset, err := graph.FindAsset("assets/test.jpg")

	assert.NoError(t, err)
	assert.Equal(t, &asset, foundAsset)
}

func TestGraph_FindAsset_CaseInsensitive(t *testing.T) {
	graph := logseq.NewGraph()
	asset := logseq.NewAsset("assets/test.jpg")
	_ = graph.AddAsset(&asset)
	foundAsset, err := graph.FindAsset("ASSETS/TEST.JPG")

	assert.NoError(t, err)
	assert.Equal(t, &asset, foundAsset)
}

func TestGraph_FindAsset_NotFound(t *testing.T) {
	graph := logseq.NewGraph()
	_, err := graph.FindAsset("assets/test.jpg")

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.AssetNotFoundError{AssetPath: "assets/test.jpg"})
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

func TestGraph_FindPage_CaseInsensitive(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage("test page")

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

func TestGraph_PublicGraph(t *testing.T) {
	graph := logseq.NewGraph()

	publicPage := logseq.NewEmptyPage()
	publicPage.Name = "Test Page"
	publicPage.PathInSite = "test-page"
	publicPage.Root.Properties.Set("public", "true")
	_ = graph.AddPage(&publicPage)

	privatePage := logseq.NewEmptyPage()
	privatePage.Name = "Private Page"
	privatePage.PathInSite = "private-page"
	privatePage.Root.Properties.Set("public", "false")
	_ = graph.AddPage(&privatePage)

	publicGraph := graph.PublicGraph()
	foundPage, err := publicGraph.FindPage("Test Page")

	assert.NoError(t, err)
	assert.Equal(t, &publicPage, foundPage)

	_, err = publicGraph.FindPage("Private Page")
	assert.Error(t, err)
}

func TestGraph_ResourceLinks(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = "Test Page"
	page.PathInSite = "test-page"
	resource, label := ExternalResource(), LinkLabel()
	link, _ := page.Root.Content.AddLinkToResource(resource, label)
	graph.AddPage(&page)
	links := graph.ResourceLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}
