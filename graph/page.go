package graph

import (
	"fmt"
	"regexp"
)

type Page struct {
	Name        string   `json:"-"`
	Title       string   `json:"title"`
	PathInGraph string   `json:"-"`
	Kind        string   `json:"kind"`
	Root        *Block   `json:"root"`
	AllBlocks   []*Block `json:"-"`
}

func NewEmptyPage() Page {
	root := NewEmptyBlock()

	return Page{
		Kind:      "page",
		Root:      root,
		AllBlocks: []*Block{root},
	}
}

// Aliases returns alternate names for this page.
func (p *Page) Aliases() []string {
	aliasesProp, ok := p.Root.Properties.Get("alias")
	if !ok {
		return []string{}
	}

	return aliasesProp.List()
}

// IsJournal returns true if the page name looks like a journal entry.
func (p *Page) IsJournal() bool {
	dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	return dateRe.MatchString(p.Name)
}

// IsPlaceholder returns true if the page is not a file on disk.
func (p *Page) IsPlaceholder() bool {
	return p.PathInGraph == ""
}

// IsPublic returns true if the page root is public.
func (p *Page) IsPublic() bool {
	return p.Root.IsPublic()
}

// IsSection returns true if the page is a section.
func (p *Page) IsSection() bool {
	return p.Kind == "section"
}

// Links returns links collected from all blocks in the page.
func (p *Page) Links() []Link {
	links := []Link{}

	for _, block := range p.AllBlocks {
		links = append(links, block.Links()...)
	}

	return links
}

// Properties returns the root block's properties.
func (p *Page) Properties() *PropertyMap {
	return p.Root.Properties
}

// RequestsHoistedNamespace returns true if page properties specify hoisting.
func (p *Page) RequestsHoistedNamespace() bool {
	hoistProp, ok := p.Root.Properties.Get("hoist-namespace")

	return ok && hoistProp.Bool()
}

// String returns a string representation of the page.
func (p *Page) String() string {
	return fmt.Sprintf("<Page: %s>", p.Name)
}

// SetRoot assign's page root block and sets AllBlocks to root's branches.
func (p *Page) SetRoot(root *Block) {
	p.Root = root
	p.AllBlocks = []*Block{}
	p.AddTree(root)
}

// AddTree adds a block and its children to the page's AllBlocks.
func (p *Page) AddTree(block *Block) {
	p.AllBlocks = append(p.AllBlocks, block)
	for _, child := range block.Children {
		p.AddTree(child)
	}
}

// Tags returns tag properties defined for the page.
func (p *Page) Tags() []string {
	return p.Root.Tags()
}
