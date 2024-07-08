package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/lpernett/godotenv"
	log "github.com/sirupsen/logrus"

	"export-logseq/logseq"
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

	graph := logseq.LoadGraph(graphDir)
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
