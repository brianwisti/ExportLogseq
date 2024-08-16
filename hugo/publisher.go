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
	RequirePublic   bool
}

const (
	folderPermissions = 0755
)

func ExportGraph(graph graph.Graph, siteDir string, requirePublic bool) error {
	log.Infof("Exporting from %s to %s", graph.GraphDir, siteDir)

	if requirePublic {
		log.Info("Only exporting public blocks")

		graph = graph.PublicGraph()
	}

	totalPages := len(graph.Pages)
	totalAssets := len(graph.Assets)
	log.Infof("Graph has %d pages and %d assets", totalPages, totalAssets)

	exporter := Exporter{
		Graph:           graph,
		SiteDir:         siteDir,
		AssetDir:        filepath.Join(siteDir, "assets", "graph-assets"),
		ContentDir:      filepath.Join(siteDir, "content"),
		PagePermalinks:  map[string]string{},
		AssetPermalinks: map[string]string{},
		RequirePublic:   requirePublic,
	}

	exporter.PagePermalinks = exporter.SetPagePermalinks()
	exporter.AssetPermalinks = exporter.SetAssetPermalinks()

	exportedPageCount, err := exporter.ExportGraphJSON()
	if err != nil {
		return errors.Wrap(err, "exporting graph JSON")
	}

	log.Infof("Exported pages into graph JSON: %d", exportedPageCount)

	exportedPageCount, err = exporter.ExportPages()
	if err != nil {
		return errors.Wrap(err, "exporting pages")
	}

	log.Infof("Exported %d pages as pages", exportedPageCount)
	log.Infof("Exporting assets from %d asset links", len(graph.AssetLinks()))

	exportedAssetCount, err := exporter.ExportAssets()
	if err != nil {
		return errors.Wrap(err, "exporting assets")
	}

	log.Infof("Exported %d pages and %d assets", exportedPageCount, exportedAssetCount)

	return nil
}

// ExportAssets exports graph asset files to the site directory.
func (e *Exporter) ExportAssets() (int, error) {
	exportCount := 0

	log.Infof("Exporting assets to: %s", e.AssetDir)

	if err := os.MkdirAll(e.AssetDir, folderPermissions); err != nil {
		return exportCount, errors.Wrap(err, "creating asset directory "+e.AssetDir)
	}

	for _, asset := range e.Graph.Assets {
		log.Debug("Exporting asset " + asset.Name)
		targetPath := e.PublishedAssetPath(asset.Name)
		sourcePath := filepath.Join(e.Graph.GraphDir, "assets", asset.Name)
		shouldExport, err := e.ShouldExportGraphFile(sourcePath, targetPath)

		if err != nil {
			return exportCount, errors.Wrap(err, "checking if asset should be exported")
		}

		if !shouldExport {
			continue
		}

		if err := e.ExportGraphFile(sourcePath, targetPath); err != nil {
			return exportCount, errors.Wrap(err, "exporting asset "+asset.Name)
		}

		exportCount++
	}

	return exportCount, nil
}

// ExportGraphJSON exports the graph to a JSON file in the site directory.
func (e *Exporter) ExportGraphJSON() (int, error) {
	exportCount := 0
	exportDir := e.SiteDir

	if err := os.MkdirAll(exportDir, folderPermissions); err != nil {
		return exportCount, errors.Wrap(err, "creating data export directory"+exportDir)
	}

	exportDataPath := filepath.Join(exportDir, "logseq.json")
	exportFile, err := os.Create(exportDataPath)

	if err != nil {
		return exportCount, errors.Wrap(err, "creating export file")
	}

	defer exportFile.Close()

	enc := json.NewEncoder(exportFile)
	enc.SetIndent("", "  ")

	if err := enc.Encode(e.Graph); err != nil {
		return exportCount, errors.Wrap(err, "encoding graph to JSON")
	}

	exportCount = len(e.Graph.Pages)

	return exportCount, nil
}

// ExportPages exports the graph pages to the site directory.
func (e *Exporter) ExportPages() (int, error) {
	log.Info("Exporting pages to: ", e.ContentDir)

	exportCount := 0

	if err := os.MkdirAll(e.ContentDir, folderPermissions); err != nil {
		return exportCount, errors.Wrap(err, "creating content directory "+e.ContentDir)
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

				return
			}

			exportCount++
		}(wg, page)
	}

	wg.Wait()

	return exportCount, nil
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

	if pageContent == "" {
		log.Warn("No content found for page: ", page.Name)
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

func (e *Exporter) ShouldSkipBlock(block graph.Block) bool {
	if e.RequirePublic && !block.IsPublic() {
		log.Warn("Skipping non-public block: ", block.String())

		return true
	}

	if block.IsTask() {
		log.Debug("Skipping task block: ", block.String())

		return true
	}

	return false
}

func (e *Exporter) ConstructBlockShortcode(block graph.Block) string {
	shortcodeArgs := map[string]string{}

	shortcodeArgs["id"] = block.ID

	if block.Depth > 0 {
		captionProp, ok := block.Properties.Get("caption")
		if ok {
			shortcodeArgs["caption"] = strings.Replace(captionProp.Value, "\"", "\\\"", -1)
		}

		callout := block.Callout()
		if callout != "" {
			shortcodeArgs["callout"] = callout
		}

		if !block.IsPublic() {
			shortcodeArgs["classes"] = "private"
		}
	}

	shortCode := "block"

	for arg, value := range shortcodeArgs {
		shortCode = shortCode + " " + arg + "=\"" + value + "\""
	}

	return shortCode
}

// ProcessBlock turns a block and its children into Hugo content.
func (e *Exporter) ProcessBlock(block graph.Block) (string, error) {
	log.Debug("Processing block ", block.ID)

	if e.ShouldSkipBlock(block) {
		return "", nil
	}

	blockContent := ""

	if block.Depth > 0 {
		// root block technically has no content. Only process children.
		blockContent = block.Content.Markdown

		if block.Content.IsCodeBlock() {
			blockContent = strings.Replace(blockContent, "{{<", "{{/**/<", -1)
		} else {
			blockContent = e.ProcessBlockEmbeddedShortcodes(blockContent)
		}

		// process page links
		for _, link := range block.Links() {
			replacement := e.ProcessBlockLink(link)
			blockContent = strings.Replace(blockContent, link.Raw, replacement, -1)
		}
	}

	shortCode := e.ConstructBlockShortcode(block)

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
		permalink, ok := e.PermalinkForPage(link.LinkPath)
		if ok {
			return `{{< page-link link="` + permalink + `" >}}` + link.Label + "{{< /page-link >}}"
		}

		return UnavailableLink(link.Label)
	}

	if link.LinkType == graph.LinkTypeBlock {
		targetBlock, ok := e.Graph.Blocks[link.LinkPath]
		if !ok {
			log.Warn("Block not found for link: ", link.LinkPath)

			return UnavailableLink(link.Label)
		}

		blockContent := targetBlock.Content.Markdown
		permalink := e.PermalinkForBlock(*targetBlock)

		return `{{< block-link link="` + permalink + `" >}}` + blockContent + "{{< /block-link >}}"
	}

	if link.LinkType == graph.LinkTypeTag {
		permalink, ok := e.PermalinkForPage(link.LinkPath)
		if ok {
			// oops that space from identifying tags
			shortCode := fmt.Sprintf(` {{< logseq/tag-link label="%s" link="%s" >}}`, link.Label, permalink)

			return shortCode
		}

		return UnavailableLink(link.Label)
	}

	if link.LinkType == graph.LinkTypeAsset {
		permalink, ok := e.PermalinkForAsset(link.LinkPath)
		if ok {
			return "![" + link.Label + "](" + permalink + ")"
		}

		return UnavailableLink(link.Label)
	}

	return link.Label
}

func (e *Exporter) ProcessBlockEmbeddedShortcodes(blockContent string) string {
	// ex: {{video https://www.youtube.com/watch?v=0Uc3ZrmhDN4}}
	videoRe := regexp.MustCompile(`\{\{video (?:(?:https://www\.youtube\.com/watch\?v=)|(?:https://youtu.be/))([^}]+)\}\}`)

	embeddedVideo := `<div
	style="position: relative; padding-bottom: 56.25%; height: 0; overflow: hidden;">
	<iframe
		allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
		allowfullscreen="allowfullscreen"
		loading="eager"
		referrerpolicy="strict-origin-when-cross-origin"
		src="https://www.youtube.com/embed/$1?autoplay=0&controls=1&end=0&loop=0&mute=0&start=0" 
		style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; border:0;" 
		title="YouTube video"
	></iframe>
	</div>`
	blockContent = videoRe.ReplaceAllString(blockContent, embeddedVideo)

	// ex: {{renderer :linkpreview,https://golangci-lint.run/usage/configuration/}}
	linkPreviewRe := regexp.MustCompile(`\{\{renderer :linkpreview,([^}]+)\}\}`)
	blockContent = linkPreviewRe.ReplaceAllString(blockContent, "$1")

	return blockContent
}

// UnavailableLink returns a string used to indicate a missing link.
func UnavailableLink(label string) string {
	return "*" + label + "*"
}

func (e *Exporter) ExportGraphFile(sourcePath string, targetPath string) error {
	log.Debugf("Exporting asset: %s → %s", sourcePath, targetPath)

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

func (e *Exporter) determinePageFrontmatter(page graph.Page) string {
	date := ""
	backlinks := []string{}
	banner := ""
	summary := ""
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

		pagePermalink, ok := e.PermalinkForPage(block.PageName)

		if !ok {
			log.Warn("No permalink found for block: ", blockID)

			continue
		}

		permalink := pagePermalink + "#" + block.ID
		backlinks = append(backlinks, permalink)
	}

	for _, tagLink := range e.Graph.FindTagLinksToPage(&page) {
		blockID := tagLink.LinksFrom

		log.Debug("Found tag link from: ", blockID)

		block, ok := e.Graph.Blocks[blockID]

		if !ok {
			log.Warn("Block not found for tag link: ", blockID)

			continue
		}

		permalink, ok := e.PermalinkForPage(block.PageName)

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
			tagPermalink, ok := e.PermalinkForPage(tagKey)

			if !ok {
				log.Debug("No permalink found for tag: ", tag)

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

		banner, ok = e.PermalinkForAsset(bannerPath)
		if !ok {
			log.Warn("No permalink found for banner asset: ", bannerPath)
		}
	}

	summaryProp, ok := page.Root.Properties.Get("summary")
	if ok {
		summary = summaryProp.String()
	}

	frontmatter := struct {
		Title     string   `json:"title"`
		Date      string   `json:"date,omitempty"`
		Backlinks []string `json:"backlinks,omitempty"`
		Tags      []string `json:"tags,omitempty"`
		TagLinks  []string `json:"taglinks,omitempty"`
		Banner    string   `json:"banner,omitempty"`
		Summary   string   `json:"summary,omitempty"`
	}{
		Title:     page.Title,
		Date:      date,
		Backlinks: backlinks,
		Tags:      tagList,
		TagLinks:  tagLinks,
		Banner:    banner,
		Summary:   summary,
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
	// Contents is a special case, serving as site root
	permalinks["contents"] = "/"

	for _, page := range e.Graph.Pages {
		if page.Name == "Contents" {
			continue
		}

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

// PermalinkForAsset determines the permalink for an Asset.
func (e *Exporter) PermalinkForAsset(assetName string) (string, bool) {
	nameKey := strings.ToLower(assetName)
	permalink, ok := e.AssetPermalinks[nameKey]

	if !ok {
		log.Debug("No permalink found for asset:", assetName)
	}

	return permalink, ok
}

// PermalinkForPage determines the permalink for a Page.
func (e *Exporter) PermalinkForPage(pageName string) (string, bool) {
	nameKey := strings.ToLower(pageName)
	permalink, ok := e.PagePermalinks[nameKey]

	if !ok {
		log.Debug("No permalink found for page:", pageName)
	}

	return "/" + permalink, ok
}

// PermalinkForBlock determines the permalink for a Block.
func (e *Exporter) PermalinkForBlock(block graph.Block) string {
	pagePermalink, ok := e.PermalinkForPage(block.PageName)

	if !ok {
		log.Warn("No permalink found for block page: ", block.PageName)

		return ""
	}

	return pagePermalink + "#" + block.ID
}

// PageContentPath determines the content file path for a Page.
func (e *Exporter) PageContentPath(page graph.Page) string {
	if page.Name == "contents" {
		return filepath.Join(e.ContentDir, "_index.md")
	}

	permalink, ok := e.PermalinkForPage(page.Name)
	if !ok {
		log.Fatalf("No permalink found for page: %s", page.Name)
	}

	// Determine the target path for the page.
	pageSubtree := strings.Split(permalink, "/")
	pageSubtree = append([]string{e.ContentDir}, pageSubtree...)

	// Find pages in the page's namespace
	subpages := e.Graph.PagesInNamespace(page.Name)

	if len(subpages) > 0 || permalink == "/" {
		pageSubtree = append(pageSubtree, "_index")
	}

	contentPath := filepath.Join(pageSubtree...) + ".md"

	return contentPath
}

func (e *Exporter) PublishedAssetPath(assetName string) string {
	assetBase := filepath.Base(assetName)

	return filepath.Join(e.AssetDir, assetBase)
}

func (e *Exporter) ShouldExportGraphFile(sourcePath string, targetPath string) (bool, error) {
	fileExt := filepath.Ext(sourcePath)

	if fileExt == ".pdf" {
		log.Info("Skip PDF asset: ", sourcePath)

		return false, nil
	}

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
	}

	if targetFileStat == nil {
		log.Debugf("Target file does not exist: %s", targetPath)

		return true, nil
	}

	if !targetFileStat.Mode().IsRegular() {
		return false, errors.Errorf("target file is not a regular file: %s", targetPath)
	}

	if os.SameFile(sourceFileStat, targetFileStat) {
		log.Debugf("source and target are the same file: %s", sourcePath)

		return false, nil
	}

	sourceModTime := sourceFileStat.ModTime()
	targetModTime := targetFileStat.ModTime()

	if sourceModTime.After(targetModTime) {
		log.Debugf("source is newer than target: %s", sourcePath)

		return true, nil
	}

	return false, nil
}
