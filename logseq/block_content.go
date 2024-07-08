package logseq

import "regexp"

type BlockContent struct {
	Block         *Block  `json:"block"` // Block that contains this content
	Markdown      string  `json:"markdown"`
	HTML          string  `json:"html"`
	PageLinks     []*Link `json:"page_links"`
	ResourceLinks []*Link `json:"resource_links"`
}

func EmptyBlockContent() *BlockContent {
	return &BlockContent{}
}

func BlockContentFromRawSource(block *Block, rawSource string) *BlockContent {
	content := BlockContent{
		Block:    block,
		Markdown: rawSource,
	}

	content.findResourceLinks()
	content.findPageLinks()

	return &content
}

func (bc *BlockContent) findPageLinks() {
	pageLinkRe := regexp.MustCompile(`\[\[(.*?)\]\]`)
	pageLinks := []*Link{}

	for _, match := range pageLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		raw, pageName := match[0], match[1]
		// This is more of a bookmark, to be replaced with an actual Page link by the graph.
		link := Link{
			Raw:       raw,
			LinksFrom: bc.Block,
			LinksTo:   &Page{Name: pageName},
			Label:     pageName,
			IsEmbed:   false,
		}
		pageLinks = append(pageLinks, &link)
	}

	bc.PageLinks = pageLinks
}

func (bc *BlockContent) findResourceLinks() {
	resourceLinkRe := regexp.MustCompile(`(!?)\[(.*?)\]\((.*?)\)`)
	resourceLinks := []*Link{}

	for _, match := range resourceLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		raw, isEmbed, resourceTitle, resourceUrl := match[0], match[1], match[2], match[3]
		resource := ExternalResource{Uri: resourceUrl}
		link := Link{
			Raw:       raw,
			LinksFrom: bc.Block,
			LinksTo:   resource,
			Label:     resourceTitle,
			IsEmbed:   isEmbed == "!",
		}

		resourceLinks = append(resourceLinks, &link)
	}

	bc.ResourceLinks = resourceLinks
}
