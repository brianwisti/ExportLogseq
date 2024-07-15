package graph_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"export-logseq/graph"
)

func TestGraph_NewGraph(t *testing.T) {
	graph := graph.NewGraph()

	assert.NotNil(t, graph)
	assert.Empty(t, graph.Pages)
	assert.Empty(t, graph.Assets)
}

func TestGraph_AddAsset(t *testing.T) {
	g := graph.NewGraph()
	asset := graph.NewAsset("assets/test.jpg")
	err := g.AddAsset(&asset)

	assert.NoError(t, err)

	assetKey := strings.ToLower(asset.PathInGraph)
	addedAsset, ok := g.Assets[assetKey]

	assert.True(t, ok)
	assert.Equal(t, &asset, addedAsset)
}

func TestGraph_AddAsset_WithExistingAsset(t *testing.T) {
	g := graph.NewGraph()
	asset := graph.NewAsset("assets/test.jpg")
	_ = g.AddAsset(&asset)
	err := g.AddAsset(&asset)

	assert.Error(t, err)
	assert.ErrorIs(t, err, graph.AssetExistsError{AssetPath: asset.PathInGraph})
}

func TestGraph_AddPage(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	err := g.AddPage(&page)
	assert.NoError(t, err)

	pageKey := strings.ToLower(page.Name)
	addedPage, ok := g.Pages[pageKey]

	assert.True(t, ok)
	assert.Equal(t, &page, addedPage)
}

func TestGraph_AddPage_WithExistingPage(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	_ = g.AddPage(&page)
	err := g.AddPage(&page)

	assert.Error(t, err)
	assert.ErrorIs(t, err, graph.PageExistsError{PageName: page.Name})
}

func TestGraph_FindAsset(t *testing.T) {
	g := graph.NewGraph()
	asset := graph.NewAsset("assets/test.jpg")
	_ = g.AddAsset(&asset)
	foundAsset, ok := g.FindAsset("assets/test.jpg")

	assert.True(t, ok)
	assert.Equal(t, &asset, foundAsset)
}

func TestGraph_FindAsset_CaseInsensitive(t *testing.T) {
	g := graph.NewGraph()
	assetName := "assets/test.jpg"
	asset := graph.NewAsset(assetName)
	_ = g.AddAsset(&asset)
	foundAsset, ok := g.FindAsset(strings.ToUpper(assetName))

	assert.False(t, ok)
	assert.Nil(t, foundAsset)
}

func TestGraph_FindAsset_NotFound(t *testing.T) {
	graph := graph.NewGraph()
	_, ok := graph.FindAsset("assets/test.jpg")

	assert.False(t, ok)
}

func TestGraph_FindLinksToPage(t *testing.T) {
	g := graph.NewGraph()
	fromPage, toPage := Page(), Page()
	err := g.AddPage(&fromPage)
	require.NoError(t, err)

	err = g.AddPage(&toPage)
	require.NoError(t, err)

	link := graph.Link{
		LinkPath: toPage.Name,
		Label:    toPage.Name,
		LinkType: graph.LinkTypePage,
		IsEmbed:  false,
	}

	link, err = fromPage.Root.Content.AddLink(link)
	require.NoError(t, err)

	links := g.FindLinksToPage(&toPage)
	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_FindPage(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	_ = g.AddPage(&page)
	foundPage, err := g.FindPage(page.Name)

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_FindPage_CaseInsensitive(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	_ = g.AddPage(&page)
	foundPage, err := g.FindPage(page.Name)

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_FindPage_NotFound(t *testing.T) {
	g := graph.NewGraph()
	_, err := g.FindPage("Test Page")

	assert.Error(t, err)
	assert.ErrorIs(t, err, graph.PageNotFoundError{PageName: "Test Page"})
}

func TestGraph_FindPage_WithAlias(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.Root.Properties.Set("alias", "alias")
	_ = g.AddPage(&page)
	foundPage, err := g.FindPage("alias")

	assert.NoError(t, err)
	assert.Equal(t, &page, foundPage)
}

func TestGraph_PagesInNamespace(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = "test/page"
	_ = g.AddPage(&page)
	pages := g.PagesInNamespace("test")

	assert.NotEmpty(t, pages)
	assert.Contains(t, pages, &page)
}

func TestGraph_Links(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	link := graph.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: graph.LinkTypeResource,
		IsEmbed:  false,
	}
	link, _ = page.Root.Content.AddLink(link)
	g.AddPage(&page)
	links := g.Links()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_Links_Empty(t *testing.T) {
	graph := graph.NewGraph()
	links := graph.Links()

	assert.Empty(t, links)
}

func TestGraph_AssetLinks(t *testing.T) {
	g := graph.NewGraph()
	page := Page()
	g.AddPage(&page)

	asset := graph.NewAsset("assets/test.jpg")
	_ = g.AddAsset(&asset)
	link := graph.Link{
		LinkPath: asset.PathInGraph,
		Label:    asset.PathInGraph,
		LinkType: graph.LinkTypeAsset,
		IsEmbed:  false,
	}
	link, _ = page.Root.Content.AddLink(link)
	links := g.AssetLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_PageLinks_Empty(t *testing.T) {
	graph := graph.NewGraph()
	links := graph.PageLinks()

	assert.Empty(t, links)
}

func TestGraph_PageLinks(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()

	g.AddPage(&page)

	pageName, label := PageName(), LinkLabel()
	link := graph.Link{
		LinkPath: pageName,
		Label:    label,
		LinkType: graph.LinkTypePage,
		IsEmbed:  false,
	}

	link, _ = page.Root.Content.AddLink(link)

	links := g.PageLinks()
	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_ResourceLinks(t *testing.T) {
	g := graph.NewGraph()
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	url, label := gofakeit.URL(), gofakeit.Phrase()
	link := graph.Link{
		LinkPath: url,
		Label:    label,
		LinkType: graph.LinkTypeResource,
	}
	link, _ = page.Root.Content.AddLink(link)
	g.AddPage(&page)
	links := g.ResourceLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}

func TestGraph_PublicGraph(t *testing.T) {
	g := graph.NewGraph()

	publicPage := graph.NewEmptyPage()
	publicPage.Name = gofakeit.Word()
	publicPage.Root.Properties.Set("public", "true")
	_ = g.AddPage(&publicPage)

	privatePage := graph.NewEmptyPage()
	privatePage.Name = "Private Page"
	privatePage.Root.Properties.Set("public", "false")
	_ = g.AddPage(&privatePage)

	publicGraph := g.PublicGraph()
	foundPage, err := publicGraph.FindPage(publicPage.Name)

	assert.NoError(t, err)
	assert.Equal(t, &publicPage, foundPage)

	_, err = publicGraph.FindPage("Private Page")
	assert.Error(t, err)
}

func TestGraph_PublicGraph_WithAssetInPublicPage(t *testing.T) {
	g := graph.NewGraph()

	publicPage := graph.NewEmptyPage()
	publicPage.Name = gofakeit.Word()
	publicPage.Root.Properties.Set("public", "true")
	_ = g.AddPage(&publicPage)

	asset := graph.NewAsset("assets/test.jpg")
	_ = g.AddAsset(&asset)
	link := graph.Link{
		LinkPath: asset.PathInGraph,
		Label:    asset.PathInGraph,
		LinkType: graph.LinkTypeAsset,
		IsEmbed:  false,
	}
	link, _ = publicPage.Root.Content.AddLink(link)

	publicGraph := g.PublicGraph()
	links := publicGraph.AssetLinks()

	assert.NotEmpty(t, links)
	assert.Contains(t, links, link)
}
