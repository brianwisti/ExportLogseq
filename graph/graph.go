package graph

import (
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Graph struct {
	GraphDir          string            `json:"-"`
	Name              string            `json:"name"`
	HoistedNamespaces []string          `json:"-"`
	Pages             map[string]*Page  `json:"pages"`
	Blocks            map[string]*Block `json:"-"`
	Assets            []Asset           `json:"assets"`
	links             []Link
}

func NewGraph() Graph {
	return Graph{
		Pages:             map[string]*Page{},
		Assets:            []Asset{},
		Blocks:            map[string]*Block{},
		HoistedNamespaces: []string{},
	}
}

// AddAsset adds an asset to the graph.
func (g *Graph) AddAsset(asset Asset) error {
	_, assetExists := g.FindAsset(asset.Path)

	if assetExists {
		return AssetExistsError{asset.Name}
	}

	log.Debug("Adding asset" + asset.Name)
	g.Assets = append(g.Assets, asset)

	return nil
}

// Add a single page to the graph.
func (g *Graph) AddPage(page *Page) error {
	pageKey := strings.ToLower(page.Name)
	log.Debug("Adding page with key: ", pageKey)

	// Ignore placeholder properties like "-"
	if pageKey == "-" {
		return nil
	}

	existingPage, _ := g.Pages[pageKey]

	if existingPage != nil {
		if existingPage.IsPlaceholder() {
			log.Debug("Replacing placeholder page: ", page.Name)
		} else {
			return PageExistsError{page.Name}
		}
	}

	g.Pages[pageKey] = page

	for _, tag := range page.Tags() {
		tagKey := strings.ToLower(tag)
		_, ok := g.Pages[tagKey]

		if !ok {
			_, err := g.AddPlaceholderPage(tag)
			if err != nil {
				return errors.Wrap(err, "adding tag page "+tag)
			}
		}
	}

	for _, tagLink := range page.TagLinks() {
		tagKey := strings.ToLower(tagLink.LinkPath)
		_, ok := g.Pages[tagKey]

		if !ok {
			_, err := g.AddPlaceholderPage(tagLink.LinkPath)
			if err != nil {
				return errors.Wrap(err, "adding tag page "+tagLink.LinkPath)
			}
		}
	}

	for _, block := range page.AllBlocks {
		g.Blocks[block.ID] = block
	}

	if page.RequestsHoistedNamespace() {
		g.HoistedNamespaces = append(g.HoistedNamespaces, page.Name)
	}

	return nil
}

// AddPlaceholderPage adds a placeholder page to the graph.
// Placeholder pages are used to represent pages that do not exist in the graph.
func (g *Graph) AddPlaceholderPage(name string) (*Page, error) {
	pageKey := strings.ToLower(name)
	_, pageExists := g.Pages[pageKey]

	if pageExists {
		return nil, PageExistsError{name}
	}

	page := NewEmptyPage()
	nameSteps := strings.Split(name, "/")
	title := nameSteps[len(nameSteps)-1]

	page.Name = name
	page.Title = title

	page.Root.Properties.Set("public", "true")

	return &page, g.AddPage(&page)
}

// PageIsHoisted returns true if the page is in a hoisted namespace.
func (g *Graph) PageIsHoisted(page *Page) bool {
	for _, ns := range g.HoistedNamespaces {
		if page.Name == ns || strings.HasPrefix(page.Name, ns+"/") {
			return true
		}
	}

	return false
}

// FindAsset returns an asset by path.
func (g *Graph) FindAsset(path string) (Asset, bool) {
	for _, asset := range g.Assets {
		if asset.Path == path {
			return asset, true
		}
	}

	return Asset{}, false
}

// FindLinksToPage returns all links to a Page.
func (g *Graph) FindLinksToPage(page *Page) []Link {
	log.Debug("Finding links in graph to: ", page)

	links := []Link{}

	for _, link := range g.PageLinks() {
		linkTarget := link.LinkPath
		log.Debug("Checking link from ", link.LinksFrom, " to ", linkTarget)

		if page.Name == linkTarget {
			log.Debug("Found link to ", page.Name, " in ", link.LinksFrom)
			links = append(links, link)
		}
	}

	return links
}

// FindTagLinksToPage returns all tag links to a Page.
func (g *Graph) FindTagLinksToPage(page *Page) []Link {
	log.Debug("Finding tag links in graph to: ", page)

	links := []Link{}

	for _, link := range g.Links() {
		if link.LinkType != LinkTypeTag {
			continue
		}

		linkTarget := link.LinkPath
		log.Debug("Checking link from ", link.LinksFrom, " to ", linkTarget)

		if page.Name == linkTarget {
			log.Debug("Found link to ", page.Name, " in ", link.LinksFrom)
			links = append(links, link)
		}
	}

	return links
}

// FindPage returns a page by name or alias.
func (g *Graph) FindPage(name string) (*Page, error) {
	pageKey := strings.ToLower(name)
	page, ok := g.Pages[pageKey]

	if ok {
		return page, nil
	}

	for _, page := range g.Pages {
		for _, alias := range page.Aliases() {
			if alias == name {
				return page, nil
			}
		}
	}

	return nil, PageNotFoundError{name}
}

// Links returns all links found in the graph.
func (g *Graph) Links() []Link {
	if g.links == nil {
		links := []Link{}
		for _, page := range g.Pages {
			links = append(links, page.Links()...)
		}

		g.links = links
	}

	return g.links
}

// AssetLinks returns all asset links found in the graph.
func (g *Graph) AssetLinks() []Link {
	links := []Link{}

	for _, link := range g.Links() {
		if link.LinkType == LinkTypeAsset {
			links = append(links, link)
		}
	}

	return links
}

// PageLinks returns all page links found in the graph.
func (g *Graph) PageLinks() []Link {
	links := []Link{}

	for _, link := range g.Links() {
		if link.LinkType == LinkTypePage {
			links = append(links, link)
		}
	}

	return links
}

// PagesInNamespace returns all pages in a namespace.
func (g *Graph) PagesInNamespace(namespace string) []*Page {
	pages := []*Page{}
	asPrefix := namespace + "/"

	for _, page := range g.Pages {
		if strings.HasPrefix(page.Name, asPrefix) {
			pages = append(pages, page)
		}
	}

	return pages
}

// PublicGraph returns a copy of the graph with only public pages.
func (g *Graph) PublicGraph() Graph {
	publicGraph := NewGraph()
	publicGraph.GraphDir = g.GraphDir
	publicGraph.Name = g.Name

	for _, page := range g.Pages {
		if page.IsPublic() {
			publicGraph.AddPage(page)
		}
	}

	// Add assets that are linked from public pages.
	for _, link := range publicGraph.AssetLinks() {
		if link.LinkType == LinkTypeAsset {
			log.Debugf("Checking asset %s", link.LinkPath)
			asset, ok := g.FindAsset(link.LinkPath)

			if !ok {
				log.Fatalf("Asset not in original graph: [%s]", link.LinkPath)
			}

			_, ok = publicGraph.FindAsset(link.LinkPath)
			if !ok {
				err := publicGraph.AddAsset(asset)
				if err != nil {
					log.Fatalf("adding asset %s to public graph: %v", link.LinkPath, err)
				}
			}
		}
	}

	return publicGraph
}

// ResourceLinks returns all resource links found in the graph.
func (g *Graph) ResourceLinks() []Link {
	links := []Link{}

	for _, link := range g.Links() {
		if link.LinkType == LinkTypeResource {
			links = append(links, link)
		}
	}

	log.Debug("Resource links found: ", len(links))

	return links
}
