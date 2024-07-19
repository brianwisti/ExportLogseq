package hugo

import (
	"encoding/json"
	"export-logseq/graph"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	Graph           graph.Graph
	SiteDir         string
	AssetDir        string
	ContentDir      string
	PagePermalinks  map[string]string
	AssetPermalinks map[string]string
}

const (
	folderPermissions = 0755
)

func ExportGraph(graph graph.Graph, siteDir string) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)
	exporter := Exporter{
		Graph:           graph,
		SiteDir:         siteDir,
		AssetDir:        filepath.Join(siteDir, "assets", "graph-assets"),
		ContentDir:      filepath.Join(siteDir, "content"),
		PagePermalinks:  map[string]string{},
		AssetPermalinks: map[string]string{},
	}

	exporter.PagePermalinks = exporter.SetPagePermalinks()
	exporter.AssetPermalinks = exporter.SetAssetPermalinks()

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
	log.Warning("Removing existing asset directory")

	if err := os.RemoveAll(e.AssetDir); err != nil {
		return errors.Wrap(err, "removing existing asset directory")
	}

	if err := os.MkdirAll(e.AssetDir, folderPermissions); err != nil {
		return errors.Wrap(err, "creating asset directory "+e.AssetDir)
	}

	for _, asset := range e.Graph.Assets {
		log.Debug("Exporting asset " + asset.Name)

		if err := e.ExportLinkedAsset(*asset); err != nil {
			return errors.Wrap(err, "exporting asset "+asset.Name)
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

	blockContent := ""

	if block.Depth > 0 {
		// root block technically has no content. Only process children.
		blockContent = block.Content.Markdown
		// process page links
		for _, link := range block.Links() {
			replacement := e.ProcessBlockLink(link)
			blockContent = strings.Replace(blockContent, link.Raw, replacement, -1)
		}

		blockContent = strings.Replace(blockContent, "{{<", "{{/**/<", -1)
	}

	shortcodeArgs := map[string]string{}

	shortcodeArgs["id"] = block.ID

	captionProp, ok := block.Properties.Get("caption")
	if ok {
		shortcodeArgs["caption"] = strings.Replace(captionProp.Value, "\"", "\\\"", -1)
	}

	shortCode := "block"

	for arg, value := range shortcodeArgs {
		shortCode = shortCode + " " + arg + "=\"" + value + "\""
	}

	if block.IsHeader() {
		headerString := fmt.Sprintf(`{{%% block-header level=%d %%}}%s{{%% /block-header %%}}`, block.Depth, blockContent)
		blockContent = headerString
	}

	processedContent := "\n{{% " + shortCode + " %}}" + blockContent

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

func (e *Exporter) ProcessBlockLink(link graph.Link) string {
	if link.LinkType == graph.LinkTypePage {
		permalink, ok := e.PagePermalink(link.LinkPath)
		if ok {
			return "[" + link.Label + "](" + permalink + ")"
		}

		return UnavailableLink(link.Label)
	}

	if link.LinkType == graph.LinkTypeAsset {
		permalink, ok := e.AssetPermalink(link.LinkPath)
		if ok {
			return "![" + link.Label + "](" + permalink + ")"
		}

		return UnavailableLink(link.Label)
	}

	return link.Label
}

// UnavailableLink returns a string used to indicate a missing link.
func UnavailableLink(label string) string {
	return "*" + label + "*"
}

func (e *Exporter) ExportLinkedAsset(asset graph.Asset) error {
	targetPath := e.PublishedAssetPath(asset.Name)
	sourcePath := filepath.Join(e.Graph.GraphDir, "assets", asset.PathInGraph)

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
	banner := ""
	tagList := []string{}
	tagLinks := []string{}

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

	for _, tagLink := range e.Graph.FindTagLinksToPage(&page) {
		blockID := tagLink.LinksFrom

		log.Debug("Found tag link from: ", blockID)

		block, ok := e.Graph.Blocks[blockID]

		if !ok {
			log.Warn("Block not found for tag link: ", blockID)

			continue
		}

		permalink, ok := e.PagePermalink(block.PageName)

		if !ok {
			log.Warn("No permalink found for block: ", blockID)

			continue
		}

		tagLink := "[" + block.PageName + "](" + permalink + ")"
		tagLinks = append(tagLinks, tagLink)
	}

	tagsProp, ok := page.Root.Properties.Get("tags")
	if ok {
		tags := tagsProp.List()

		log.Debug("Found tags property: ", tags)

		for _, tag := range tags {
			tagKey := strings.ToLower(tag)
			tagPermalink, ok := e.PagePermalink(tagKey)

			if !ok {
				log.Warn("No permalink found for tag: ", tag)

				continue
			}

			tagLink := "[" + tag + "](" + tagPermalink + ")"
			tagList = append(tagList, tagLink)
		}
	}

	dateProp, ok := page.Root.Properties.Get("date")
	if ok {
		date = dateProp.String()
	} else if page.IsJournal() {
		date = page.Name
	}

	bannerProp, ok := page.Root.Properties.Get("banner")
	if ok {
		bannerPath := strings.TrimPrefix(bannerProp.String(), "../assets/")
		log.Debug("Found banner property: ", bannerPath)

		banner, ok = e.AssetPermalink(bannerPath)
		if !ok {
			log.Warn("No permalink found for banner asset: ", bannerPath)
		}
	}

	frontmatter := struct {
		Title     string   `json:"title"`
		Date      string   `json:"date,omitempty"`
		Backlinks []string `json:"backlinks,omitempty"`
		Tags      []string `json:"tags,omitempty"`
		TagLinks  []string `json:"taglinks,omitempty"`
		Banner    string   `json:"banner,omitempty"`
	}{
		Title:     page.Title,
		Date:      date,
		Backlinks: backlinks,
		Tags:      tagList,
		TagLinks:  tagLinks,
		Banner:    banner,
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

// SetAssetPermalinks builds a map of asset names to permalinks.
func (e *Exporter) SetAssetPermalinks() map[string]string {
	permalinks := map[string]string{}

	for _, asset := range e.Graph.Assets {
		nameKey := strings.ToLower(asset.Name)
		permalinks[nameKey] = "/graph-assets/" + asset.Name
	}

	return permalinks
}

// SetPagePermalinks builds a map of page names to permalinks.
func (e *Exporter) SetPagePermalinks() map[string]string {
	permalinks := map[string]string{}

	for _, page := range e.Graph.Pages {
		nameSteps := strings.Split(page.Name, "/")
		slugSteps := []string{}

		if !e.Graph.PageIsHoisted(page) {
			section := "pages"
			dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

			if dateRe.MatchString(page.Name) {
				section = "journals"
			}

			slugSteps = append(slugSteps, section)
		}

		for _, step := range nameSteps {
			slugSteps = append(slugSteps, slug.Make(step))
		}

		permalink := strings.Join(slugSteps, "/")
		nameKey := strings.ToLower(page.Name)
		permalinks[nameKey] = permalink

		// Don't forget page aliases!
		for _, alias := range page.Aliases() {
			aliasKey := strings.ToLower(alias)
			permalinks[aliasKey] = permalink
		}
	}

	return permalinks
}

// AssetPermalink determines the permalink for an Asset.
func (e *Exporter) AssetPermalink(assetName string) (string, bool) {
	nameKey := strings.ToLower(assetName)
	permalink, ok := e.AssetPermalinks[nameKey]

	if !ok {
		log.Debug("No permalink found for asset:", assetName)
	}

	return permalink, ok
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

func (e *Exporter) PublishedAssetPath(assetName string) string {
	assetBase := filepath.Base(assetName)

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
	if err != nil {
		if !os.IsNotExist(err) {
			return false, errors.Wrap(err, "checking target file")
		}

		return true, nil
	}

	if !targetFileStat.Mode().IsRegular() {
		return false, errors.Errorf("target file is not a regular file: %s", targetPath)
	}

	if os.SameFile(sourceFileStat, targetFileStat) {
		log.Debugf("source and target are the same file: %s", sourcePath)

		return false, nil
	}

	return true, nil
}
