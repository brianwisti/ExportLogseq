package logseq

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/wikilink"

	"export-logseq/graph"
)

const (
	branchBlockOpener    = "- "
	branchBlockContinuer = "  "
)

type GraphLinkResolver struct {
	Graph graph.Graph
}

// Loader loads a Logseq graph from a directory.
type Loader struct {
	GraphDir string
	Graph    graph.Graph
}

func NewLoader(graphDir string) Loader {
	g := graph.NewGraph()
	g.GraphDir = graphDir
	g.Name = filepath.Base(graphDir)
	log.Info("Graph name: ", g.Name)

	return Loader{GraphDir: graphDir, Graph: g}
}

// LoadGraph loads a Logseq graph from a directory.
func LoadGraph(graphDir string) (graph.Graph, error) {
	log.Info("Loading Logseq graph from", graphDir)
	loader := NewLoader(graphDir)

	if err := loader.loadAssets(); err != nil {
		return loader.Graph, errors.Wrap(err, "loading assets")
	}

	pageDirs := []string{"pages", "journals"}

	for _, pageDir := range pageDirs {
		if err := loader.loadPagesFromDir(pageDir); err != nil {
			return loader.Graph, errors.Wrap(err, "loading pages from "+pageDir)
		}
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, &wikilink.Extender{}),
	)

	for _, block := range loader.Graph.Blocks {
		var buf bytes.Buffer

		if err := md.Convert([]byte(block.Content.Markdown), &buf); err != nil {
			log.Fatal("converting markdown to HTML:", err)
		}

		block.Content.HTML = buf.String()
	}

	// Gather backlinks and tag links to pages
	for _, page := range loader.Graph.Pages {
		page.Backlinks = loader.Graph.FindLinksToPage(page)
		page.TaggedLinks = loader.Graph.FindTagLinksToPage(page)
	}

	return loader.Graph, nil
}

func (loader *Loader) LoadPage(pageFile string, graphPath string) (graph.Page, error) {
	baseName := filepath.Base(pageFile)
	namespace := "pages"
	fullPageFileName := strings.ReplaceAll(baseName, "___", "/")
	fullPageName := strings.TrimSuffix(fullPageFileName, ".md")
	title := fullPageName
	journalDateRe := regexp.MustCompile(`^\d{4}_\d{2}_\d{2}$`)
	titleIsJournalDate := journalDateRe.MatchString(fullPageName)

	if titleIsJournalDate {
		fullPageName = strings.Replace(fullPageName, "_", "-", -1)
		namespace = "journals"
	} else {
		escapedName, decodeErr := url.QueryUnescape(fullPageName)
		if decodeErr != nil {
			return graph.Page{}, errors.New("decoding page name: " + decodeErr.Error())
		}

		fullPageName = escapedName
	}

	pathInGraph, err := filepath.Rel(graphPath, pageFile)
	if err != nil {
		return graph.Page{}, errors.Wrap(err, "calculating path in graph")
	}

	nameSteps := strings.Split(fullPageName, "/")
	stepCount := len(nameSteps)

	if stepCount > 1 {
		title = nameSteps[len(nameSteps)-1]
		namespace = namespace + "/" + strings.Join(nameSteps[:stepCount-1], "/")
	}

	// Generate a slug for the page name
	slugs := []string{}

	for _, step := range nameSteps {
		slugs = append(slugs, slug.Make(step))
	}

	slugName := strings.Join(slugs, "/")

	// Process each line of fullPageName
	file, err := os.Open(pageFile)
	if err != nil {
		return graph.Page{}, errors.New("opening page file: " + err.Error())
	}
	defer file.Close()

	lines, err := LoadPageLines(file)
	if err != nil {
		return graph.Page{}, errors.New("loading page lines: " + err.Error())
	}

	page := graph.Page{
		Name:        fullPageName,
		Title:       title,
		Namespace:   namespace,
		PathInGraph: pathInGraph,
		Path:        slugName,
	}

	blocks, err := findBlocks(&page, lines)
	if err != nil {
		return graph.Page{}, errors.New("finding blocks: " + err.Error())
	}

	if len(blocks) == 0 {
		log.Warn("No root block found in page: ", fullPageName)

		blocks = []*graph.Block{graph.NewEmptyBlock()}
	}

	page.Root = blocks[0]
	page.AllBlocks = blocks

	return page, nil
}

func (loader *Loader) loadAssets() error {
	assetsDir := filepath.Join(loader.GraphDir, "assets")
	log.Info("Assets directory:", assetsDir)
	assetFiles, err := filepath.Glob(filepath.Join(assetsDir, "*.*"))

	if err != nil {
		return errors.Wrap(err, "listing asset files")
	}

	for _, assetFile := range assetFiles {
		relPath, err := filepath.Rel(assetsDir, assetFile)
		if err != nil {
			return errors.Wrap(err, "calculating relative path for asset")
		}

		asset := graph.NewAsset(relPath)
		err = loader.Graph.AddAsset(asset)

		if err != nil {
			return errors.Wrap(err, "adding asset "+assetFile)
		}
	}

	return nil
}

func (loader *Loader) loadPagesFromDir(subdir string) error {
	g := &loader.Graph
	pagesDir := filepath.Join(g.GraphDir, subdir)
	log.Infof("Loading pages from %s", pagesDir)
	pageFiles, err := filepath.Glob(filepath.Join(pagesDir, "*.md"))

	if err != nil {
		return errors.Wrap(err, "listing page files")
	}

	wg := new(sync.WaitGroup)
	pageCh := make(chan graph.Page, len(pageFiles))
	errCh := make(chan error, 1)

	for _, pageFile := range pageFiles {
		if filepath.Base(pageFile) == "Templates.md" {
			continue
		}

		wg.Add(1)

		go func(wg *sync.WaitGroup, pageFile string) {
			defer wg.Done()

			page, err := loader.LoadPage(pageFile, pagesDir)
			if err != nil {
				errCh <- errors.Wrap(err, "loading page "+pageFile)

				return
			}

			pageCh <- page
		}(wg, pageFile)
	}

	wg.Wait()
	close(errCh)
	close(pageCh)

	err = <-errCh
	if err != nil {
		return errors.Wrap(err, "loading pages")
	}

	for page := range pageCh {
		err := g.AddPage(&page)
		if err != nil {
			return errors.Wrap(err, "adding page "+page.Name)
		}
	}

	return nil
}

func findBlocks(page *graph.Page, lines []PageLine) ([]*graph.Block, error) {
	blocks := []*graph.Block{}
	blockStack := NewBlockStack()
	currentBlockLines := []string{}
	currentIndent := 0

	for _, line := range lines {
		log.Debug("Line: ", line)
		// Skip empty block lines
		if line.Content == "-" {
			continue
		}

		if strings.HasPrefix(line.Content, branchBlockOpener) {
			// Remember the current block.
			block, err := graph.NewBlock(page, currentBlockLines, currentIndent)

			if err != nil {
				return nil, errors.Wrap(err, "creating new block")
			}

			blocks = append(blocks, block)
			blockStack = PlaceBlock(block, blockStack)

			// Adjust for the root block not having a branch block marker.
			line.Indent = line.Indent + 1

			// Reset the current block and indent
			currentBlockLines = []string{}
			currentIndent = line.Indent
			line.Content = strings.TrimPrefix(line.Content, branchBlockOpener)
		} else if strings.HasPrefix(line.Content, branchBlockContinuer) {
			// Ensure that the current line is a continuation of a current block
			if len(currentBlockLines) == 0 {
				return blocks, errors.New("no block to continue: " + line.Content)
			}

			line.Content = strings.TrimPrefix(line.Content, branchBlockContinuer)
			// Adjust for the root block not having a branch block marker.
			line.Indent = line.Indent + 1
		}

		// Ensure that the current line is indented correctly
		if line.Indent != currentIndent {
			errMsg := fmt.Sprintf("mismatched indent: %v", line)

			return blocks, errors.New(errMsg)
		}

		currentBlockLines = append(currentBlockLines, line.Content)
	}

	// Remember the last block.
	if len(currentBlockLines) > 0 {
		block, err := graph.NewBlock(page, currentBlockLines, currentIndent)

		if err != nil {
			return nil, errors.Wrap(err, "creating block from remaining lines")
		}

		blocks = append(blocks, block)
		PlaceBlock(block, blockStack)
	}

	log.Debug("Blocks: ", blocks)

	return blocks, nil
}
