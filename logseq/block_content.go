package logseq

import (
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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
	content.SetMarkdown(rawSource)

	return content
}

// FindLinkToPage checks if the block content already links to the page.
func (bc *BlockContent) FindLinkToPage(pageName string) *Link {
	for _, link := range bc.PageLinks {
		target := link.LinksTo.(*Page)
		if target.Name == pageName {
			return link
		}
	}

	return nil
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

// AddLinkToPage adds a link to a page to the block content.
func (bc *BlockContent) AddLinkToPage(pageName string, label string) (*Link, error) {
	log.Debugf("Adding link to page: label=%s, target=%s", label, pageName)

	existingLink := bc.FindLinkToPage(pageName)
	if existingLink != nil {
		return nil, ErrorDuplicatePageLink{PageName: pageName}
	}

	// This is more of a bookmark, to be replaced with an actual Page link during InContext.
	target := NewEmptyPage()
	target.Name = pageName
	link := Link{
		LinksFrom: bc.Block,
		LinksTo:   &target,
		Label:     label,
		IsEmbed:   false,
	}

	bc.PageLinks = append(bc.PageLinks, &link)

	return &link, nil
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

// SetMarkdown sets the markdown content of the block.
func (bc *BlockContent) SetMarkdown(markdown string) error {
	bc.Markdown = markdown

	err := bc.findLinks()
	if err != nil {
		return errors.Wrap(err, "finding links")
	}

	return nil
}

func (bc *BlockContent) findLinks() error {
	bc.findPageLinks()

	err := bc.findResourceLinks()
	if err != nil {
		return errors.Wrap(err, "finding resource links")
	}

	return nil
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

func (bc *BlockContent) findResourceLinks() error {
	resourceLinkRe := regexp.MustCompile(`(!?)\[(.*?)\]\((.*?)\)`)
	resourceLinks := []*Link{}

	for _, match := range resourceLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		isEmbed, label, resourceUrl := match[1], match[2], match[3]
		log.Debug("Found resource link: ", match[0], isEmbed, label, resourceUrl)
		resource := ExternalResource{Uri: resourceUrl}

		if isEmbed == "!" {
			_, err := bc.AddEmbeddedLinkToResource(resource, label)
			if err != nil {
				return errors.Wrap(err, "adding embedded link to resource")
			}

			continue
		}

		_, err := bc.AddLinkToResource(resource, label)
		if err != nil {
			return errors.Wrap(err, "adding link to resource")
		}
	}

	bc.ResourceLinks = resourceLinks

	return nil
}
