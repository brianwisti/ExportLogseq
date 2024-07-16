package hugo

import (
	"encoding/json"
	"export-logseq/graph"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	Graph          graph.Graph
	SiteDir        string
	AssetDir       string
	ContentDir     string
	PagePermalinks map[string]string
}

const (
	folderPermissions = 0755
)

func ExportGraph(graph graph.Graph, siteDir string) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)
	exporter := Exporter{
		Graph:          graph,
		SiteDir:        siteDir,
		AssetDir:       filepath.Join(siteDir, "static", "img"),
		ContentDir:     filepath.Join(siteDir, "content"),
		PagePermalinks: map[string]string{},
	}

	exporter.PagePermalinks = exporter.SetPagePermalinks()

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
	log.Warning("Removing existing content directory")

	if err := os.RemoveAll(e.AssetDir); err != nil {
		return errors.Wrap(err, "removing existing asset directory")
	}

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

	wg := new(sync.WaitGroup)
	c := make(chan int)

	for _, page := range e.Graph.Pages {
		wg.Add(1)

		go func(wg *sync.WaitGroup, page *graph.Page) {
			defer wg.Done()

			err := e.exportPage(*page)

			if err != nil {
				log.Error("Error exporting page:", err)
				close(c)
			}
		}(wg, page)
	}

	wg.Wait()

	return nil
}

// ProcessBlock turns a block and its children into Hugo content.
func (e *Exporter) ProcessBlock(block graph.Block) (string, error) {
	log.Debug("Processing block ", block.ID)

	if !block.IsPublic() {
		log.Debug("Skipping non-public block ", block.ID)

		return "", nil
	}

	blockContent := block.Content.Markdown

	// process page links
	for _, link := range block.Links() {
		if link.LinkType == graph.LinkTypePage {
			replacement := "*" + link.Label + "*"

			permalink, ok := e.PagePermalink(link.LinkPath)
			if ok {
				replacement = "[" + link.Label + "](" + permalink + ")"
			}

			blockContent = strings.Replace(blockContent, link.Raw, replacement, -1)
		}
	}

	blockContent = strings.Replace(blockContent, "{{<", "{{/**/<", -1)
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
	log.Debug("Exporting page:", page.Name)

	contentPath := e.PageContentPath(page)
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
	date := ""
	backlinks := []string{}

	for _, link := range e.Graph.FindLinksToPage(&page) {
		blockID := link.LinksFrom

		log.Debug("Found backlink from: ", blockID)

		block, ok := e.Graph.Blocks[blockID]

		if !ok {
			log.Warn("Block not found for link: ", blockID)

			continue
		}

		pagePermalink, ok := e.PagePermalink(block.PageName)

		if !ok {
			log.Warn("No permalink found for block: ", blockID)

			continue
		}

		backlink := "[" + block.PageName + "](" + pagePermalink + ")"

		backlinks = append(backlinks, backlink)
	}

	dateProp, ok := page.Root.Properties.Get("date")
	if ok {
		date = dateProp.String()
	}

	frontmatter := struct {
		Title     string   `json:"title"`
		Date      string   `json:"date"`
		Backlinks []string `json:"backlinks"`
	}{
		Title:     page.Title,
		Date:      date,
		Backlinks: backlinks,
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

// SetPagePermalinks builds a map of page names to permalinks.
func (e *Exporter) SetPagePermalinks() map[string]string {
	permalinks := map[string]string{}

	for _, page := range e.Graph.Pages {
		nameSteps := strings.Split(page.Name, "/")
		slugSteps := []string{}

		for _, step := range nameSteps {
			slugSteps = append(slugSteps, slug.Make(step))
		}

		permalink := strings.Join(slugSteps, "/")
		nameKey := strings.ToLower(page.Name)
		permalinks[nameKey] = permalink
	}

	return permalinks
}

// PagePermalink determines the permalink for a Page.
func (e *Exporter) PagePermalink(pageName string) (string, bool) {
	nameKey := strings.ToLower(pageName)
	permalink, ok := e.PagePermalinks[nameKey]

	if !ok {
		log.Debug("No permalink found for page:", pageName)
	}

	return "/" + permalink, ok
}

// PageContentPath determines the content file path for a Page.
func (e *Exporter) PageContentPath(page graph.Page) string {
	permalink, ok := e.PagePermalink(page.Name)
	if !ok {
		log.Fatalf("No permalink found for page: %s", page.Name)
	}

	// Determine the target path for the page.
	pageSubtree := strings.Split(permalink, "/")
	pageSubtree = append([]string{e.ContentDir}, pageSubtree...)

	// Find pages in the page's namespace
	subpages := e.Graph.PagesInNamespace(page.Name)

	if len(subpages) > 0 {
		pageSubtree = append(pageSubtree, "_index")
	}

	contentPath := filepath.Join(pageSubtree...) + ".md"

	return contentPath
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
