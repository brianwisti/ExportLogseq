package logseq

import (
	"bytes"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
)

type Graph struct {
	GraphDir   string
	Pages      map[string]*Page `json:"pages"`
	AssetPaths []string         `json:"-"`
}

func LoadGraph(graphDir string) *Graph {
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
	assetPaths := []string{}
	assetFiles, err := filepath.Glob(filepath.Join(assetsDir, "*.*"))
	if err != nil {
		log.Fatal("listing asset files:", err)
	}

	for _, assetFile := range assetFiles {
		relPath, err := filepath.Rel(assetsDir, assetFile)
		if err != nil {
			log.Fatal("calculating relative path for asset:", err)
		}

		assetPaths = append(assetPaths, relPath)
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

	pages := map[string]*Page{}

	for _, pageFile := range pageFiles {
		page, err := LoadPage(pageFile, pagesDir)
		if err != nil {
			log.Fatalf("loading page %s: %v", pageFile, err)
		}
		pages[page.Name] = &page
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
		pages[page.Name] = &page
	}

	return &Graph{
		GraphDir:   graphDir,
		Pages:      pages,
		AssetPaths: assetPaths,
	}
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

func (g *Graph) prepPageForSite(page *Page) {
	blockCount := len(page.AllBlocks)
	log.Debug("Prepping ", page.Name, " with ", blockCount, " blocks")

	for _, block := range page.AllBlocks {
		g.prepBlockForSite(block)
	}
}

func (g *Graph) prepBlockForSite(block *Block) {
	blockMarkdown := block.Content.Markdown
	log.Debug("Prepping block ", block.ID, " with ", len(block.Content.PageLinks), " page links")
	log.Debug("Initial block Markdown: ", blockMarkdown)
	for i := 0; i < len(block.Content.PageLinks); i++ {
		link := block.Content.PageLinks[i]
		log.Debug("Raw link: ", link.Raw)
		linkTarget := link.LinksTo.(*Page)

		if targetPage, ok := g.Pages[linkTarget.Name]; ok {
			// update the link to point to the actual page if needed
			if linkTarget != targetPage {
				link.LinksTo = targetPage
			}

			permalink, err := targetPage.InContext(*g)
			if err != nil {
				log.Fatal("getting permalink for ", targetPage.Name, ":", err)
			}

			log.Debug("Linking ", block.ID, " to ", permalink)
			mdLink := "[" + link.Label + "](" + permalink + ")"
			blockMarkdown = strings.Replace(blockMarkdown, link.Raw, mdLink, 1)
		}
	}

	log.Debug("Final block Markdown: ", blockMarkdown)

	block.Content.Markdown = blockMarkdown

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(block.Content.Markdown), &buf); err != nil {
		log.Fatal("converting markdown to HTML:", err)
	}
	block.Content.HTML = buf.String()
}
