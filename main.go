package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lpernett/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"

	"export-logseq/internal/logseq"
	config "export-logseq/internal/logseq"
)

type Graph struct {
	Pages map[string]*logseq.Page `json:"pages"`
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

func (g *Graph) prepPageForSite(page *logseq.Page) {
	log.Debug("Assigning links for ", page.Name)
	for _, block := range page.Blocks {
		g.prepBlockForSite(block)
	}
}

func (g *Graph) prepBlockForSite(block *logseq.Block) {
	for i := 0; i < len(block.Content.PageLinks); i++ {
		link := block.Content.PageLinks[i]
		if targetPage, ok := g.Pages[link.Url]; ok {
			permalink := "/" + targetPage.PathInSite
			log.Debug("Linking ", block.ID, " to ", permalink)
			mdLink := "[" + link.Title + "](" + permalink + ")"
			block.Content.Markdown = strings.Replace(block.Content.Markdown, link.Raw, mdLink, -1)
		}
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(block.Content.Markdown), &buf); err != nil {
		log.Fatal("converting markdown to HTML:", err)
	}
	block.Content.HTML = buf.String()

	for _, childBlock := range block.Children {
		g.prepBlockForSite(childBlock)
	}
}
func main() {
	start := time.Now()
	log.Info("Initializing...")
	dotEnvErr := godotenv.Load()
	if dotEnvErr != nil {
		log.Fatal("Loading .env file:", dotEnvErr)
	}

	graphDir := os.Getenv("GRAPH_DIR")
	log.Info("GRAPH_DIR:", graphDir)
	if graphDir == "" {
		log.Fatal("GRAPH_DIR is not set in .env file or environment variables")
	}

	siteDir := os.Getenv("SITE_DIR")
	log.Info("SITE_DIR:", siteDir)
	if graphDir == "" {
		log.Fatal("SITE_DIR is not set in .env file or environment variables")
	}

	configFile := filepath.Join(graphDir, "logseq", "config.edn")
	logseqConfig, err := config.LoadConfig(configFile)
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

	pagesDir := filepath.Join(graphDir, "pages")
	log.Info("Pages directory:", pagesDir)
	pageFiles, err := filepath.Glob(filepath.Join(pagesDir, "*.md"))
	if err != nil {
		log.Fatal("listing page files:", err)
	}

	pages := map[string]*logseq.Page{}

	for _, pageFile := range pageFiles {
		page, err := logseq.LoadPage(pageFile, pagesDir)
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
		page, err := logseq.LoadPage(journalFile, journalsDir)
		if err != nil {
			log.Fatalf("loading journal %s: %v", journalFile, err)
		}
		pages[page.Name] = &page
	}

	graph := Graph{Pages: pages}
	graph.PutPagesInContext()

	exportPath := filepath.Join(siteDir, "logseq.json")
	exportFile, err := os.Create(exportPath)
	if err != nil {
		log.Fatal("creating export file:", err)
	}
	defer exportFile.Close()

	enc := json.NewEncoder(exportFile)
	enc.SetIndent("", "  ")
	enc.Encode(graph)
	log.Info("All done!")
	elapsed := time.Since(start)
	log.Infof("Elapsed time: %s", elapsed)
}
