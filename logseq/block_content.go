package logseq

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// BlockContent represents the content of a block in a Logseq graph.
type BlockContent struct {
	BlockID  string          `json:"block_id"` // Block that contains this content
	Markdown string          `json:"markdown"`
	HTML     string          `json:"html"`
	Links    map[string]Link `json:"links"`
	Callout  string          `json:"callout"`
}

func NewEmptyBlockContent() *BlockContent {
	return &BlockContent{
		BlockID:  "",
		Markdown: "",
		HTML:     "",
		Callout:  "",
		Links:    map[string]Link{},
	}
}

func NewBlockContent(block *Block, rawSource string) *BlockContent {
	content := NewEmptyBlockContent()
	content.BlockID = block.ID
	err := content.SetMarkdown(rawSource)

	if err != nil {
		log.Errorf("Error setting markdown content: %s", err)
	}

	return content
}

// AddLink adds a link to the block content.
func (bc *BlockContent) AddLink(link Link) (Link, error) {
	log.Debugf("Adding link from block %s: %s", bc.BlockID, link.LinkPath)

	_, ok := bc.FindLink(link.LinkPath)
	if ok {
		return Link{}, ErrorDuplicateLink{link.LinkPath}
	}

	link.LinksFrom = bc.BlockID
	bc.Links[link.LinkPath] = link

	return link, nil
}

// FindLink returns a link by path.
func (bc *BlockContent) FindLink(path string) (Link, bool) {
	link, ok := bc.Links[path]

	return link, ok
}

func (bc *BlockContent) IsCodeBlock() bool {
	codeBlockRe := regexp.MustCompile("```")

	return codeBlockRe.MatchString(bc.Markdown)
}

// SetMarkdown sets the markdown content of the block.
func (bc *BlockContent) SetMarkdown(markdown string) error {
	calloutRe := regexp.MustCompile(`(?sm)#\+BEGIN_(\S+)\n(.+?)\n#\+END_(\S+)`)
	calloutMatch := calloutRe.FindStringSubmatch(markdown)

	if calloutMatch != nil {
		opener, body, closer := calloutMatch[1], calloutMatch[2], calloutMatch[3]

		if opener != closer {
			log.Fatalf("(%s) callout mismatch: %s != %s", bc.BlockID, opener, closer)
		}

		bc.Callout = strings.ToLower(opener)
		log.Debugf("(%s) found callout: %s", bc.BlockID, bc.Callout)

		markdown = calloutRe.ReplaceAllString(markdown, body)
		log.Debug("New Markdown: ", markdown)
	}

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
		raw, pageName := match[0], match[1]
		log.Debugf("Found page link: [%s] -> %s", raw, pageName)
		link := Link{
			Raw:       raw,
			LinksFrom: bc.BlockID,
			LinkPath:  pageName,
			Label:     pageName,
			LinkType:  LinkTypePage,
			IsEmbed:   false,
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
		raw, isEmbed, label, resourceUrl, isAsset := match[0], match[1], match[2], match[3], match[4]
		log.Debugf("Found resource link: ->%s<- label=%s uri=%s", raw, label, resourceUrl)

		linkType := LinkTypeResource

		if isAsset != "" {
			linkType = LinkTypeAsset
			resourceUrl = strings.TrimPrefix(resourceUrl, "..")
		}

		link := Link{
			Raw:       raw,
			LinksFrom: bc.BlockID,
			LinkPath:  resourceUrl,
			Label:     label,
			LinkType:  linkType,
			IsEmbed:   isEmbed == "!",
		}

		_, err := bc.AddLink(link)
		if err != nil {
			return errors.Wrap(err, "adding resource link")
		}
	}

	return nil
}
