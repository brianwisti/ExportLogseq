package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"export-logseq/graph"
	"export-logseq/logseq"
)

type SelectedPages string

const (
	AllPages    SelectedPages = "all"
	PublicPages SelectedPages = "public"
)

type ExportCmd struct {
	GraphDir      string        `env:"GRAPH_DIR" arg:""            help:"Path to the Logseq graph directory."`
	SiteDir       string        `env:"SITE_DIR" arg:""            help:"Path to the site directory."`
	SelectedPages SelectedPages `enum:"all,public" default:"public" help:"Select pages to export."`
}

func (cmd *ExportCmd) Run() error {
	graph, err := logseq.LoadGraph(cmd.GraphDir)

	if err != nil {
		return errors.Wrap(err, "loading graph")
	}

	if cmd.SelectedPages == PublicPages {
		graph = graph.PublicGraph()
	}

	if err := exportGraph(&graph, cmd.SiteDir); err != nil {
		return errors.Wrap(err, "exporting graph")
	}

	return nil
}

type CLI struct {
	Export ExportCmd `cmd:"" help:"Export a Logseq graph to SSG content folder."`
}

func exportGraph(graph *graph.Graph, siteDir string) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)

	exportDir := filepath.Join(siteDir, "assets", "exported")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return errors.Wrap(err, "creating data export directory"+exportDir)
	}

	exportDataPath := filepath.Join(exportDir, "logseq.json")
	exportFile, err := os.Create(exportDataPath)

	if err != nil {
		return errors.Wrap(err, "creating export file")
	}

	defer exportFile.Close()

	enc := json.NewEncoder(exportFile)
	enc.SetIndent("", "  ")

	if err := enc.Encode(graph); err != nil {
		log.Fatal("encoding graph:", err)
	}

	// Export linked assets to site directory
	assetDir := filepath.Join(siteDir, "static")

	log.Infof("Exporting assets to: %s", assetDir)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return errors.Wrap(err, "creating asset directory "+assetDir)
	}

	for _, link := range graph.AssetLinks() {
		asset, ok := graph.FindAsset(link.LinkPath)

		if !ok {
			return errors.Errorf("asset not found: %s", link.LinkPath)
		}

		targetPath := filepath.Join(assetDir, asset.PathInSite)
		sourcePath := filepath.Join(graph.GraphDir, link.LinkPath)
		targetDir := filepath.Dir(targetPath)
		log.Debug("Exporting asset:", sourcePath, "→", targetDir)

		if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
			return errors.Wrap(err, "creating target directory for assets")
		}

		log.Debugf("Exporting asset: %s → %s", sourcePath, targetPath)
		// Copy the file at sourcePath to targetPath
		sourceFileStat, err := os.Stat(sourcePath)
		if err != nil {
			return errors.Wrap(err, "getting source file info")
		}

		if !sourceFileStat.Mode().IsRegular() {
			return errors.Errorf("source file is not a regular file: %s", sourcePath)
		}

		targetFileStat, err := os.Stat(targetPath)
		if err == nil {
			if !targetFileStat.Mode().IsRegular() {
				return errors.Errorf("target file is not a regular file: %s", targetPath)
			}

			if os.SameFile(sourceFileStat, targetFileStat) {
				log.Debugf("source and target are the same file: %s", sourcePath)

				continue
			}
		} else {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "checking target file")
			}
		}

		source, err := os.Open(sourcePath)
		if err != nil {
			return errors.Wrap(err, "opening source file")
		}

		defer source.Close()

		target, err := os.Create(targetPath)
		if err != nil {
			return errors.Wrap(err, "creating target file")
		}

		defer target.Close()

		if _, err := io.Copy(target, source); err != nil {
			return errors.Wrap(err, "copying file")
		}
	}

	pageCount := len(graph.Pages)
	log.Infof("Exported %d pages to: %s", pageCount, exportDataPath)

	return nil
}

func main() {
	log.SetLevel(log.InfoLevel)
	log.Info("Initializing...")

	var cli CLI

	start := time.Now()
	ctx := kong.Parse(&cli)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)

	elapsed := time.Since(start)

	log.Info("All done!")
	log.Infof("Elapsed time: %s", elapsed)
}
