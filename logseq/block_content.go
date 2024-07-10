package logseq

import (
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// BlockContent represents the content of a block in a Logseq graph.
type BlockContent struct {
	Block    *Block           `json:"block"` // Block that contains this content
	Markdown string           `json:"markdown"`
	HTML     string           `json:"html"`
	Links    map[string]*Link `json:"links"`
}

func NewEmptyBlockContent() *BlockContent {
	return &BlockContent{
		Links: map[string]*Link{},
	}
}

func NewBlockContent(block *Block, rawSource string) *BlockContent {
	content := NewEmptyBlockContent()
	content.Block = block
	content.SetMarkdown(rawSource)

	return content
}

// AddLink adds a link to the block content.
func (bc *BlockContent) AddLink(link Link) (Link, error) {
	log.Debugf("Adding link from block %s: %s", bc.Block, link.LinkPath)

	_, ok := bc.FindLink(link.LinkPath)
	if ok {
		return Link{}, ErrorDuplicateLink{link.LinkPath}
	}

	link.LinksFrom = bc.Block
	bc.Links[link.LinkPath] = &link

	return link, nil
}

// FindLink returns a link by path.
func (bc *BlockContent) FindLink(path string) (*Link, bool) {
	link, ok := bc.Links[path]

	return link, ok
}

func (bc *BlockContent) IsCodeBlock() bool {
	codeBlockRe := regexp.MustCompile("```")
	return codeBlockRe.MatchString(bc.Markdown)
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

	if bc.IsCodeBlock() {
		return
	}

	for _, match := range pageLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		pageName := match[1]
		log.Debugf("Found page link: [%s] -> %s", match[0], pageName)
		link := Link{
			LinkPath: pageName,
			Label:    pageName,
			LinkType: LinkTypePage,
			IsEmbed:  false,
		}
		_, err := bc.AddLink(link)
		if err != nil {
			log.Errorf("Error adding link to page: %s", err)
		}
	}
}

func (bc *BlockContent) findResourceLinks() error {
	resourceLinkRe := regexp.MustCompile(`(!?)\[(.*?)\]\(((../assets/)?.*?)\)`)

	for _, match := range resourceLinkRe.FindAllStringSubmatch(bc.Markdown, -1) {
		isEmbed, label, resourceUrl, isAsset := match[1], match[2], match[3], match[4]
		log.Debugf("Found resource link: ->%s<- label=%s uri=%s", match[0], label, resourceUrl)

		linkType := LinkTypeResource

		if isAsset != "" {
			linkType = LinkTypeAsset
		}

		link := Link{
			LinkPath: resourceUrl,
			Label:    label,
			LinkType: linkType,
			IsEmbed:  isEmbed == "!",
		}

		_, err := bc.AddLink(link)
		if err != nil {
			return errors.Wrap(err, "adding resource link")
		}
	}

	return nil
}
