package logseq_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
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
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	_ = graph.AddPage(&page)
	err := graph.AddPage(&page)

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.PageExistsError{PageName: page.Name})
}

func TestGraph_FindAsset(t *testing.T) {
	graph := logseq.NewGraph()
	asset := logseq.NewAsset("assets/test.jpg")
	_ = graph.AddAsset(&asset)
	foundAsset, ok := graph.FindAsset("assets/test.jpg")

	assert.True(t, ok)
	assert.Equal(t, &asset, foundAsset)
}

func TestGraph_FindAsset_CaseInsensitive(t *testing.T) {
	graph := logseq.NewGraph()
	assetName := "assets/test.jpg"
	asset := logseq.NewAsset(assetName)
	_ = graph.AddAsset(&asset)
	foundAsset, ok := graph.FindAsset(strings.ToUpper(assetName))

	assert.False(t, ok)
	assert.Nil(t, foundAsset)
}

func TestGraph_FindAsset_NotFound(t *testing.T) {
	graph := logseq.NewGraph()
	_, ok := graph.FindAsset("assets/test.jpg")

	assert.False(t, ok)
}

func TestGraph_FindLinksToPage(t *testing.T) {
	graph := logseq.NewGraph()
	fromPage, toPage := Page(), Page()
	err := graph.AddPage(&fromPage)
	require.NoError(t, err)

	err = graph.AddPage(&toPage)
	require.NoError(t, err)

	link := logseq.Link{
		LinkPath: toPage.Name,
		Label:    toPage.Name,
		LinkType: logseq.LinkTypePage,
		IsEmbed:  false,
	}

	link, err = fromPage.Root.Content.AddLink(link)
	require.NoError(t, err)

	links := graph.FindLinksToPage(&toPage)
	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_FindPage(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage(page.Name)

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_FindPage_CaseInsensitive(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage(page.Name)

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
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	page.Root.Properties.Set("alias", "alias")
	_ = graph.AddPage(&page)
	foundPage, err := graph.FindPage("alias")

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_Links(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	link := logseq.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: logseq.LinkTypeResource,
		IsEmbed:  false,
	}
	link, _ = page.Root.Content.AddLink(link)
	graph.AddPage(&page)
	links := graph.Links()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_Links_Empty(t *testing.T) {
	graph := logseq.NewGraph()
	links := graph.Links()

	assert.Empty(t, links)
}

func TestGraph_AssetLinks(t *testing.T) {
	graph := logseq.NewGraph()
	page := Page()
	graph.AddPage(&page)

	asset := logseq.NewAsset("assets/test.jpg")
	_ = graph.AddAsset(&asset)
	link := logseq.Link{
		LinkPath: asset.PathInGraph,
		Label:    asset.PathInGraph,
		LinkType: logseq.LinkTypeAsset,
		IsEmbed:  false,
	}
	link, _ = page.Root.Content.AddLink(link)
	links := graph.AssetLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_PageLinks_Empty(t *testing.T) {
	graph := logseq.NewGraph()
	links := graph.PageLinks()

	assert.Empty(t, links)
}

func TestGraph_PageLinks(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()

	graph.AddPage(&page)

	pageName, label := PageName(), LinkLabel()
	link := logseq.Link{
		LinkPath: pageName,
		Label:    label,
		LinkType: logseq.LinkTypePage,
		IsEmbed:  false,
	}

	link, _ = page.Root.Content.AddLink(link)

	links := graph.PageLinks()
	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_ResourceLinks(t *testing.T) {
	graph := logseq.NewGraph()
	page := logseq.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	url, label := gofakeit.URL(), gofakeit.Phrase()
	link := logseq.Link{
		LinkPath: url,
		Label:    label,
		LinkType: logseq.LinkTypeResource,
	}
	link, _ = page.Root.Content.AddLink(link)
	graph.AddPage(&page)
	links := graph.ResourceLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_PublicGraph(t *testing.T) {
	graph := logseq.NewGraph()

	publicPage := logseq.NewEmptyPage()
	publicPage.Name = gofakeit.Word()
	publicPage.PathInSite = gofakeit.Word()
	publicPage.Root.Properties.Set("public", "true")
	_ = graph.AddPage(&publicPage)

	privatePage := logseq.NewEmptyPage()
	privatePage.Name = "Private Page"
	privatePage.PathInSite = "private-page"
	privatePage.Root.Properties.Set("public", "false")
	_ = graph.AddPage(&privatePage)

	publicGraph := graph.PublicGraph()
	foundPage, err := publicGraph.FindPage(publicPage.Name)

	assert.NoError(t, err)
	assert.Equal(t, &publicPage, foundPage)

	_, err = publicGraph.FindPage("Private Page")
	assert.Error(t, err)
}
