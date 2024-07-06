package logseq

import (
	"bytes"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
)

type Graph struct {
	Pages map[string]*Page `json:"pages"`
}

// Assign Page.kind of "section" based on pages whose names are prefixes of other page names.
func (g *Graph) PutPagesInContext() {
	for thisName, thisPage := range g.Pages {
		for otherName := range g.Pages {
			if thisName == otherName {
				continue
			}

			if strings.HasPrefix(otherName, thisName) {
				thisPage.Kind = "section"

				break
			}
		}

		g.prepPageForSite(thisPage)
	}
}

func (g *Graph) prepPageForSite(page *Page) {
	log.Debug("Assigning links for ", page.Name)
	for _, block := range page.Blocks {
		g.prepBlockForSite(block)
	}
}

func (g *Graph) prepBlockForSite(block *Block) {
	for i := 0; i < len(block.Content.PageLinks); i++ {
		link := block.Content.PageLinks[i]
		if targetPage, ok := g.Pages[link.Url]; ok {
			permalink := "/" + targetPage.PathInSite
			log.Debug("Linking ", block.ID, " to ", permalink)
			mdLink := "[" + link.Title + "](" + permalink + ")"
			block.Content.Markdown = strings.Replace(block.Content.Markdown, link.Raw, mdLink, -1)
		}
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(block.Content.Markdown), &buf); err != nil {
		log.Fatal("converting markdown to HTML:", err)
	}
	block.Content.HTML = buf.String()

	for _, childBlock := range block.Children {
		g.prepBlockForSite(childBlock)
	}
}
