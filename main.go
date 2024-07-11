package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lpernett/godotenv"
	log "github.com/sirupsen/logrus"

	"export-logseq/logseq"
)

func main() {
	log.SetLevel(log.InfoLevel)
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

	graph := logseq.LoadGraph(graphDir).PublicGraph()
	graph.PutPagesInContext()

	exportPath := filepath.Join(siteDir, "logseq.json")
	exportFile, err := os.Create(exportPath)
	if err != nil {
		log.Fatal("creating export file:", err)
	}
	defer exportFile.Close()

	enc := json.NewEncoder(exportFile)
	enc.SetIndent("", "  ")
	err = enc.Encode(graph)
	if err != nil {
		log.Fatal("encoding graph:", err)
	}

	// Export linked assets to site directory
	siteRoot := filepath.Dir(filepath.Dir(siteDir))
	assetDir := filepath.Join(siteRoot, "static")
	log.Infof("Exporting assets to: %s", assetDir)
	err = os.MkdirAll(assetDir, 0755)
	if err != nil {
		log.Fatalf("creating asset directory [%s]: %v", assetDir, err)
	}

	for _, link := range graph.Links() {
		if link.IsAsset() {
			asset, ok := graph.FindAsset(link.LinkPath)
			if !ok {
				log.Fatalf("finding asset: %v", err)
			}

			targetPath := filepath.Join(assetDir, asset.PathInSite)
			sourcePath := filepath.Join(graphDir, link.LinkPath)
			log.Infof("Exporting asset: %s → %s", sourcePath, targetPath)
			// Copy the file at sourcePath to targetPath
			sourceFileStat, err := os.Stat(sourcePath)
			if err != nil {
				log.Fatalf("getting source file info: %v", err)
			}

			if !sourceFileStat.Mode().IsRegular() {
				log.Fatalf("source file is not a regular file: %s", sourcePath)
			}

			targetFileStat, err := os.Stat(targetPath)
			if err == nil {
				if !targetFileStat.Mode().IsRegular() {
					log.Fatalf("target file is not a regular file: %s", targetPath)
				}

				if os.SameFile(sourceFileStat, targetFileStat) {
					log.Debugf("source and target are the same file: %s", sourcePath)
					continue
				}
			}

			source, err := os.Open(sourcePath)
			if err != nil {
				log.Fatalf("opening source file: %v", err)
			}

			defer source.Close()

			target, err := os.Create(targetPath)
			if err != nil {
				log.Fatalf("creating target file: %v", err)
			}

			defer target.Close()

			_, err = io.Copy(target, source)
			if err != nil {
				log.Fatalf("copying file: %v", err)
			}

		}
	}

	log.Info("All done!")
	elapsed := time.Since(start)
	pageCount := len(graph.Pages)
	log.Infof("Exported %d pages to: %s", pageCount, exportPath)
	log.Infof("Elapsed time: %s", elapsed)
}
