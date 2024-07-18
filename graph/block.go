package graph

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Block struct {
	ID         string        `json:"id"`
	PageName   string        `json:"-"` // Page that contains this block
	Content    *BlockContent `json:"content"`
	Properties *PropertyMap  `json:"properties,omitempty"`
	Depth      int           `json:"depth,omitempty"`
	Parent     *Block        `json:"-"`
	Children   []*Block      `json:"children,omitempty"`
}

func NewEmptyBlock() *Block {
	content := NewEmptyBlockContent()
	block := Block{
		ID:         uuid.New().String(),
		Content:    content,
		Properties: NewPropertyMap(),
	}
	content.BlockID = block.ID

	return &block
}

func NewBlock(page *Page, sourceLines []string, depth int) (*Block, error) {
	propertyRe := regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9_-]*):: (.*)")
	contentLines := []string{}
	properties := NewPropertyMap()

	for _, line := range sourceLines {
		propertyMatch := propertyRe.FindStringSubmatch(line)
		if propertyMatch != nil {
			prop_name, prop_value := propertyMatch[1], propertyMatch[2]
			properties.Set(prop_name, prop_value)

			continue
		}

		contentLines = append(contentLines, line)
	}

	uuidString := uuid.New().String()
	idProp, _ := properties.Get("id")

	if idProp.Value != "" {
		uuidString = idProp.Value
	} else {
		properties.Set("id", uuidString)
	}

	block := Block{
		ID:         uuidString,
		PageName:   page.Name,
		Depth:      depth,
		Properties: properties,
	}
	content := strings.Join(contentLines, "\n")
	blockContent, err := NewBlockContent(&block, content)

	if err != nil {
		return nil, errors.Wrap(err, "creating block content")
	}

	block.Content = blockContent

	return &block, nil
}

func (b *Block) AddChild(child *Block) {
	b.Children = append(b.Children, child)
	child.Parent = b
}

func (b *Block) InContext(g Graph) (string, error) {
	if b.PageName == "" {
		return "Setting block context", fmt.Errorf("Block %s has no page name", b.ID)
	}

	return "/" + b.PageName + "#" + b.ID, nil
}

func (b *Block) IsHeader() bool {
	if headerProp, ok := b.Properties.Get("heading"); ok {
		return headerProp.Bool()
	}

	return false
}
func (b *Block) IsPublic() bool {
	if publicProp, ok := b.Properties.Get("public"); ok {
		return publicProp.Bool()
	}

	if b.Parent != nil {
		return b.Parent.IsPublic()
	}

	return false
}

// Links returns all links found in the block.
func (b *Block) Links() []Link {
	links := []Link{}

	banner, ok := b.Properties.Get("banner")

	if ok {
		links = append(links, Link{
			Raw:       "",
			LinksFrom: b.String(),
			LinkPath:  filepath.Base(banner.Value),
			LinkType:  LinkTypeAsset,
			IsEmbed:   true,
			Label:     "",
		})
	}

	for _, link := range b.Content.Links {
		links = append(links, link)
	}

	return links
}

func (b *Block) String() string {
	return fmt.Sprintf("%s#%s", b.PageName, b.ID)
}

func (b *Block) SetProperty(name, value string) {
	b.Properties.Set(name, value)
}

func (b *Block) Tags() []string {
	tagsProp, ok := b.Properties.Get("tags")

	if !ok {
		return []string{}
	}

	return tagsProp.List()
}
