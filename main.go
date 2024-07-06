package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/lpernett/godotenv"
	log "github.com/sirupsen/logrus"

	"export-logseq/internal/logseq"
)

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
	logseqConfig, err := logseq.LoadConfig(configFile)
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

	graph := logseq.Graph{Pages: pages}
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
