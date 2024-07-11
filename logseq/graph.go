package logseq

import (
	"bytes"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type Graph struct {
	GraphDir string
	Pages    map[string]*Page  `json:"pages"`
	Assets   map[string]*Asset `json:"assets"`
}

func NewGraph() *Graph {
	return &Graph{
		Pages:  map[string]*Page{},
		Assets: map[string]*Asset{},
	}
}

func LoadGraph(graphDir string) *Graph {
	graph := NewGraph()
	graph.GraphDir = graphDir

	configFile := filepath.Join(graphDir, "logseq", "config.edn")
	logseqConfig, err := LoadConfig(configFile)

	if err != nil {
		log.Fatal("loading Logseq config:", err)
	}

	// We're specifically catering to my graph first.
	if logseqConfig.FileNameFormat != "triple-lowbar" {
		log.Fatal("Unsupported file name format:", logseqConfig.FileNameFormat)
	}

	if logseqConfig.PreferredFormat != "markdown" {
		log.Fatal("Unsupported preferred format:", logseqConfig.PreferredFormat)
	}

	assetsDir := filepath.Join(graphDir, "assets")
	log.Info("Assets directory:", assetsDir)
	assetFiles, err := filepath.Glob(filepath.Join(assetsDir, "*.*"))

	if err != nil {
		log.Fatal("listing asset files:", err)
	}

	for _, assetFile := range assetFiles {
		relPath, err := filepath.Rel(assetsDir, assetFile)
		if err != nil {
			log.Fatal("calculating relative path for asset:", err)
		}

		asset := NewAsset("/assets/" + relPath)
		asset.PathInSite = "/img/" + relPath
		err = graph.AddAsset(&asset)

		if err != nil {
			log.Fatalf("adding asset %s: %v", assetFile, err)
		}
	}

	if err != nil {
		log.Fatal("listing asset files:", err)
	}

	pagesDir := filepath.Join(graphDir, "pages")
	log.Info("Pages directory:", pagesDir)
	pageFiles, err := filepath.Glob(filepath.Join(pagesDir, "*.md"))

	if err != nil {
		log.Fatal("listing page files:", err)
	}

	for _, pageFile := range pageFiles {
		page, err := LoadPage(pageFile, pagesDir)
		if err != nil {
			log.Fatalf("loading page %s: %v", pageFile, err)
		}

		err = graph.AddPage(&page)
		if err != nil {
			log.Fatalf("adding page %s: %v", pageFile, err)
		}
	}

	journalsDir := filepath.Join(graphDir, "journals")
	log.Info("Journals directory:", journalsDir)
	journalFiles, err := filepath.Glob(filepath.Join(journalsDir, "*.md"))

	if err != nil {
		log.Fatal("listing journal files:", err)
	}

	for _, journalFile := range journalFiles {
		page, err := LoadPage(journalFile, journalsDir)
		if err != nil {
			log.Fatalf("loading journal %s: %v", journalFile, err)
		}

		err = graph.AddPage(&page)
		if err != nil {
			log.Fatalf("adding journal %s: %v", journalFile, err)
		}
	}

	return graph
}

// AddAsset adds an asset to the graph.
func (g *Graph) AddAsset(asset *Asset) error {
	assetKey := asset.PathInGraph
	_, assetExists := g.Assets[assetKey]

	if assetExists {
		return AssetExistsError{asset.PathInGraph}
	}

	log.Debugf("Adding asset: %s", asset.PathInGraph)
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

	return nil
}

// FindAsset returns an asset by path.
func (g *Graph) FindAsset(path string) (*Asset, bool) {
	asset, ok := g.Assets[path]

	return asset, ok
}

// FindLinksToPage returns all links to a Page.
func (g *Graph) FindLinksToPage(page *Page) []Link {
	log.Info("Finding links in graph to: ", page)

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

// PublicGraph returns a copy of the graph with only public pages.
func (g *Graph) PublicGraph() *Graph {
	publicGraph := NewGraph()
	publicGraph.GraphDir = g.GraphDir

	for _, page := range g.Pages {
		if page.IsPublic() {
			publicGraph.AddPage(page)
		}
	}

	// Add assets that are linked from public pages.
	for _, link := range publicGraph.Links() {
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

// Assign Page.kind of "section" based on pages whose names are prefixes of other page names.
func (g *Graph) PutPagesInContext() {
	for thisName, thisPage := range g.Pages {
		for otherName := range g.Pages {
			if thisName == otherName {
				continue
			}

			if strings.HasPrefix(otherName, thisName) {
				thisPage.Kind = "section"

				break
			}
		}

		g.prepPageForSite(thisPage)
	}
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

func (g *Graph) prepPageForSite(page *Page) {
	blockCount := len(page.AllBlocks)
	log.Debug("Prepping ", page.Name, " with ", blockCount, " blocks")

	for _, block := range page.AllBlocks {
		g.prepBlockForSite(block)
	}
}

func (g *Graph) prepBlockForSite(block *Block) {
	blockMarkdown := block.Content.Markdown
	pageLinksFromBlock := []Link{}
	assetLinksFromBlock := []Link{}

	for _, link := range block.Content.Links {
		if link.LinkType == LinkTypePage {
			pageLinksFromBlock = append(pageLinksFromBlock, link)
		} else if link.LinkType == LinkTypeAsset {
			assetLinksFromBlock = append(assetLinksFromBlock, link)
		}
	}

	log.Debug("Prepping block ", block.ID, " with ", len(pageLinksFromBlock), " page links")
	log.Debug("Initial block Markdown: ", blockMarkdown)

	for i := 0; i < len(pageLinksFromBlock); i++ {
		link := pageLinksFromBlock[i]
		log.Debug("Raw link: ", link.Raw)

		if link.LinkPath == "" {
			// Probably a bug in link-finding logic, so log the block content.
			log.Warning("Empty link in block content: ", block.Content.Markdown)

			continue
		}

		linkString := "*" + link.Label + "*"

		targetPage, err := g.FindPage(link.LinkPath)
		if err != nil {
			if _, ok := err.(DisconnectedPageError); ok {
				log.Fatalf("Linking page: %v", err)
			} else {
				log.Debugf("Block %v placeholder link: >%v<", block.ID, link.Label)
			}
		}

		if targetPage != nil {
			permalink := "/" + targetPage.PathInSite
			log.Debug("Linking ", block.ID, " to ", permalink)
			linkString = "[" + link.Label + "](" + permalink + ")"
		}

		blockMarkdown = strings.Replace(blockMarkdown, link.Raw, linkString, 1)
	}

	for i := 0; i < len(assetLinksFromBlock); i++ {
		link := assetLinksFromBlock[i]
		log.Debug("Raw asset link: ", link.Raw)

		if link.LinkPath == "" {
			// Probably a bug in link-finding logic, so log the block content.
			log.Warning("Empty asset link in block content: ", block.Content.Markdown)

			continue
		}

		asset, ok := g.FindAsset(link.LinkPath)
		if !ok {
			log.Fatalf("Asset not found: %s", link.LinkPath)
		}

		assetPath := asset.PathInSite
		log.Debug("Linking ", block.ID, " to ", assetPath)
		linkString := "![" + link.Label + "](" + assetPath + ")"
		blockMarkdown = strings.Replace(blockMarkdown, link.Raw, linkString, 1)
	}

	log.Debug("Final block Markdown: ", blockMarkdown)
	block.Content.Markdown = blockMarkdown

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	var buf bytes.Buffer

	if err := md.Convert([]byte(block.Content.Markdown), &buf); err != nil {
		log.Fatal("converting markdown to HTML:", err)
	}

	block.Content.HTML = buf.String()
}
