package hugo

import (
	"encoding/json"
	"export-logseq/graph"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	Graph      graph.Graph
	SiteDir    string
	AssetDir   string
	ContentDir string
}

const (
	folderPermissions = 0755
)

func ExportGraph(graph graph.Graph, siteDir string) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)
	exporter := Exporter{
		Graph:      graph,
		SiteDir:    siteDir,
		AssetDir:   filepath.Join(siteDir, "static", "img"),
		ContentDir: filepath.Join(siteDir, "content"),
	}

	if err := exporter.ExportGraphJSON(); err != nil {
		return errors.Wrap(err, "exporting graph JSON")
	}

	if err := exporter.ExportPages(); err != nil {
		return errors.Wrap(err, "exporting pages")
	}

	log.Infof("Exporting assets from %d asset links", len(graph.AssetLinks()))
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

// ExportPages exports the graph pages to the site directory.
func (e *Exporter) ExportPages() error {
	log.Info("Exporting pages to: ", e.ContentDir)

	log.Warning("Removing existing content directory")
	if err := os.RemoveAll(e.ContentDir); err != nil {
		return errors.Wrap(err, "removing existing content directory")
	}

	if err := os.MkdirAll(e.ContentDir, folderPermissions); err != nil {
		return errors.Wrap(err, "creating content directory "+e.ContentDir)
	}

	for _, page := range e.Graph.Pages {
		if err := e.exportPage(*page); err != nil {
			return errors.Wrap(err, "exporting page")
		}
	}

	return nil
}

// ProcessBlock turns a block and its children into Hugo content.
func (e *Exporter) ProcessBlock(block graph.Block) (string, error) {
	log.Debug("Processing block ", block.ID)
	blockContent := strings.Replace(block.Content.Markdown, "{{<", "{{/**/<", -1)

	processedContent := "\n{{% block %}}"

	if block.IsHeader() {
		headerString := fmt.Sprintf(`{{%% block-header level=%d %%}}%s{{%% /block-header %%}}`, block.Depth, blockContent)
		processedContent = processedContent + headerString
	} else {
		processedContent = processedContent + blockContent
	}

	for _, childBlock := range block.Children {
		childContent, err := e.ProcessBlock(*childBlock)
		if err != nil {
			return "", errors.Wrap(err, "processing child block")
		}

		processedContent = processedContent + childContent
	}

	processedContent = processedContent + "{{% /block %}}"

	return processedContent, nil
}

func (e *Exporter) exportLinkedAsset(link graph.Link) error {
	_, ok := e.Graph.FindAsset(link.LinkPath)

	if !ok {
		return errors.Errorf("asset not found: %s", link.LinkPath)
	}

	targetPath := e.mapLinkPath(link)
	sourcePath := filepath.Join(e.Graph.GraphDir, link.LinkPath)
	targetDir := filepath.Dir(targetPath)
	log.Debug("Exporting asset:", sourcePath, "→", targetDir)

	if err := os.MkdirAll(targetDir, folderPermissions); err != nil {
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

func (e *Exporter) exportPage(page graph.Page) error {
	log.Info("Exporting page:", page.Name)

	permalink, contentPath := e.determinePagePaths(page)
	log.Infof("Page paths: permalink=%s; contentPath=%s", permalink, contentPath)

	pageFrontmatter := e.determinePageFrontmatter(page)
	log.Debug("Page frontmatter:", pageFrontmatter)

	pageContent, err := e.ProcessBlock(*page.Root)
	if err != nil {
		return errors.Wrap(err, "processing page content")
	}

	log.Debug("Page content:", pageContent)

	pageContentFolder := filepath.Dir(contentPath)
	if err := os.MkdirAll(pageContentFolder, folderPermissions); err != nil {
		return errors.Wrap(err, "creating page content folder")
	}

	fileContent := fmt.Sprintf("---\n%s\n---\n%s", pageFrontmatter, pageContent)
	file, err := os.Create(contentPath)
	if err != nil {
		return errors.Wrap(err, "creating content file")
	}

	defer file.Close()

	if _, err := file.WriteString(fileContent); err != nil {
		return errors.Wrap(err, "writing content to file")
	}

	return nil
}

func (e *Exporter) determinePageFrontmatter(page graph.Page) string {
	frontmatter := struct {
		Title string `json:"title"`
	}{
		Title: page.Title,
	}

	// encode the frontmatter to JSON
	// and return it as a string
	frontmatterBytes, err := json.Marshal(frontmatter)
	if err != nil {
		log.Error("Error encoding frontmatter:", err)
		return ""
	}

	return string(frontmatterBytes)
}

// determinePagePaths determines the permalink and content path for a Page.
func (e *Exporter) determinePagePaths(page graph.Page) (string, string) {
	nameSteps := strings.Split(page.Name, "/")
	slugSteps := []string{}

	for _, step := range nameSteps {
		slugSteps = append(slugSteps, slug.Make(step))
	}

	permalink := strings.Join(slugSteps, "/")

	// Determine the target path for the page.
	pageSubtree := strings.Split(permalink, "/")
	pageSubtree = append([]string{e.ContentDir}, pageSubtree...)

	// If the page is a section, write it as an _index.md file
	if page.IsSection() {
		log.Info("Page is a section " + page.Name)
		pageSubtree = append(pageSubtree, "_index")
	}

	contentPath := filepath.Join(pageSubtree...) + ".md"

	return permalink, contentPath
}

func (e *Exporter) mapLinkPath(link graph.Link) string {
	assetBase := filepath.Base(link.LinkPath)

	return filepath.Join(e.AssetDir, assetBase)
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