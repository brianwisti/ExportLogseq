package logseq

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// ErrorDuplicateResourceLink is returned when a block content already has a link to a resource and another link is added.
type ErrorDuplicateResourceLink struct {
	Resource ExternalResource
}

func (e ErrorDuplicateResourceLink) Error() string {
	return "duplicate resource link: " + e.Resource.Uri
}

// BlockContent represents the content of a block in a Logseq graph.
type BlockContent struct {
	Block         *Block  `json:"block"` // Block that contains this content
	Markdown      string  `json:"markdown"`
	HTML          string  `json:"html"`
	PageLinks     []*Link `json:"page_links"`
	ResourceLinks []*Link `json:"resource_links"`
	AssetLinks    []*Link `json:"asset_links"`
}

func NewEmptyBlockContent() *BlockContent {
	return &BlockContent{
		PageLinks:     []*Link{},
		ResourceLinks: []*Link{},
		AssetLinks:    []*Link{},
	}
}

func NewBlockContent(block *Block, rawSource string) *BlockContent {
	content := NewEmptyBlockContent()
	content.Block = block
	content.Markdown = rawSource

	content.findResourceLinks()
	content.findPageLinks()

	return content
}

// FindLinkToResource checks if the block content already links to the external resource.
func (bc *BlockContent) FindLinkToResource(resource ExternalResource) *Link {
	for _, link := range bc.ResourceLinks {
		if link.LinksTo == resource {
			return link
		}
	}

	return nil
}

func (bc *BlockContent) IsCodeBlock() bool {
	codeBlockRe := regexp.MustCompile("```")
	return codeBlockRe.MatchString(bc.Markdown)
}

// AddLinkToResource adds a link to an external resource to the block content.
func (bc *BlockContent) AddLinkToResource(resource ExternalResource, label string) (*Link, error) {
	existingLink := bc.FindLinkToResource(resource)
	if existingLink != nil {
		return nil, ErrorDuplicateResourceLink{Resource: resource}
	}

	link := Link{
		LinksFrom: bc.Block,
		LinksTo:   resource,
		Label:     label,
		IsEmbed:   false,
	}

	bc.ResourceLinks = append(bc.ResourceLinks, &link)

	return &link, nil
}

// AddEmbeddedLinkToResource adds an embedded link to an external resource.
func (bc *BlockContent) AddEmbeddedLinkToResource(resource ExternalResource, label string) (*Link, error) {
	link, err := bc.AddLinkToResource(resource, label)
	if err != nil {
		return nil, err
	}

	link.IsEmbed = true

	return link, nil
}

func (bc *BlockContent) findPageLinks() {
	pageLinkRe := regexp.MustCompile(`\[\[(.+?)\]\]`)
	pageLinks := []*Link{}

	if bc.IsCodeBlock() {
		return
	}

	for _, match := range pageLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		raw, pageName := match[0], match[1]
		// This is more of a bookmark, to be replaced with an actual Page link by the graph.
		target := NewEmptyPage()
		target.Name = pageName
		link := Link{
			Raw:       raw,
			LinksFrom: bc.Block,
			LinksTo:   &target,
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
		raw, isEmbed, label, resourceUrl := match[0], match[1], match[2], match[3]
		log.Debug("Found resource link: ", raw, isEmbed, label, resourceUrl)
		resource := ExternalResource{Uri: resourceUrl}

		if isEmbed == "!" {
			_, err := bc.AddEmbeddedLinkToResource(resource, label)
			if err != nil {
				log.Fatalf("Adding embedded link to resource: %v", err)
			}
			continue
		}

		_, err := bc.AddLinkToResource(resource, label)
		if err != nil {
			log.Fatalf("Adding link to resource: %v", err)
		}
	}

	bc.ResourceLinks = resourceLinks
}
