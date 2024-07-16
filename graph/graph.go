package graph

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type Graph struct {
	GraphDir string
	Pages    map[string]*Page  `json:"pages"`
	Blocks   map[string]*Block `json:"-"`
	Assets   map[string]*Asset `json:"assets"`
}

func NewGraph() Graph {
	return Graph{
		Pages:  map[string]*Page{},
		Assets: map[string]*Asset{},
		Blocks: map[string]*Block{},
	}
}

// AddAsset adds an asset to the graph.
func (g *Graph) AddAsset(asset *Asset) error {
	assetKey := asset.PathInGraph
	_, assetExists := g.Assets[assetKey]

	if assetExists {
		return AssetExistsError{asset.PathInGraph}
	}

	log.Debug("Adding asset" + asset.PathInGraph)
	g.Assets[assetKey] = asset

	return nil
}

// Add a single page to the graph.
func (g *Graph) AddPage(page *Page) error {
	pageKey := strings.ToLower(page.Name)
	_, pageExists := g.Pages[pageKey]

	if pageExists {
		return PageExistsError{page.Name}
	}

	g.Pages[pageKey] = page

	for _, block := range page.AllBlocks {
		g.Blocks[block.ID] = block
	}

	return nil
}

// FindAsset returns an asset by path.
func (g *Graph) FindAsset(path string) (*Asset, bool) {
	asset, ok := g.Assets[path]

	return asset, ok
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
	links := []Link{}

	for _, page := range g.Pages {
		links = append(links, page.Links()...)
	}

	return links
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

	for _, page := range g.Pages {
		if page.IsPublic() {
			publicGraph.AddPage(page)
			log.Debugf("Adding public page %s with %d links", page.Name, len(page.Links()))
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
