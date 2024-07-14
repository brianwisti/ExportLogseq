package hugo

import (
	"encoding/json"
	"export-logseq/graph"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	Graph    graph.Graph
	SiteDir  string
	AssetDir string
}

const (
	folderPermissions = 0755
)

func ExportGraph(graph graph.Graph, siteDir string) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)
	exporter := Exporter{
		Graph:    graph,
		SiteDir:  siteDir,
		AssetDir: filepath.Join(siteDir, "static"),
	}

	if err := exporter.ExportGraphJSON(); err != nil {
		return errors.Wrap(err, "exporting graph JSON")
	}

	if err := exporter.ExportAssets(); err != nil {
		return errors.Wrap(err, "exporting assets")
	}

	pageCount := len(graph.Pages)
	log.Infof("Exported %d pages and %d assets", pageCount, len(graph.Assets))

	return nil
}

// ExportAssets exports graph asset files to the site directory.
func (e *Exporter) ExportAssets() error {

	log.Infof("Exporting assets to: %s", e.AssetDir)

	if err := os.MkdirAll(e.AssetDir, folderPermissions); err != nil {
		return errors.Wrap(err, "creating asset directory "+e.AssetDir)
	}

	for _, link := range e.Graph.AssetLinks() {
		if err := e.exportLinkedAsset(link); err != nil {
			return errors.Wrap(err, "exporting linked asset")
		}
	}

	return nil
}

// ExportGraphJSON exports the graph to a JSON file in the site directory.
func (e *Exporter) ExportGraphJSON() error {
	exportDir := filepath.Join(e.SiteDir, "assets", "exported")

	if err := os.MkdirAll(exportDir, folderPermissions); err != nil {
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

	if err := enc.Encode(e.Graph); err != nil {
		return errors.Wrap(err, "encoding graph to JSON")
	}

	return nil
}

func (e *Exporter) exportLinkedAsset(link graph.Link) error {
	asset, ok := e.Graph.FindAsset(link.LinkPath)

	if !ok {
		return errors.Errorf("asset not found: %s", link.LinkPath)
	}

	targetPath := filepath.Join(e.AssetDir, asset.PathInSite)
	sourcePath := filepath.Join(e.Graph.GraphDir, link.LinkPath)
	targetDir := filepath.Dir(targetPath)
	log.Debug("Exporting asset:", sourcePath, "→", targetDir)

	if err := os.MkdirAll(filepath.Dir(targetDir), folderPermissions); err != nil {
		return errors.Wrap(err, "creating target directory for assets")
	}

	log.Debugf("Exporting asset: %s → %s", sourcePath, targetPath)
	// Copy the file at sourcePath to targetPath
	shouldExport, err := e.shouldExportAsset(sourcePath, targetPath)

	if err != nil {
		return errors.Wrap(err, "checking if asset should be exported")
	}

	if !shouldExport {
		return nil
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

	return nil
}

func (e *Exporter) shouldExportAsset(sourcePath string, targetPath string) (bool, error) {
	sourceFileStat, err := os.Stat(sourcePath)
	if err != nil {
		return false, errors.Wrap(err, "getting source file info")
	}

	if !sourceFileStat.Mode().IsRegular() {
		return false, errors.Errorf("source file is not a regular file: %s", sourcePath)
	}

	targetFileStat, err := os.Stat(targetPath)
	if err == nil {
		if !targetFileStat.Mode().IsRegular() {
			return false, errors.Errorf("target file is not a regular file: %s", targetPath)
		}

		if os.SameFile(sourceFileStat, targetFileStat) {
			log.Debugf("source and target are the same file: %s", sourcePath)

			return false, nil
		}
	} else {
		if !os.IsNotExist(err) {
			return false, errors.Wrap(err, "checking target file")
		}
	}

	return true, nil
}
