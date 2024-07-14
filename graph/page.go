package graph

import (
	"fmt"
)

type Page struct {
	Name        string   `json:"-"`
	PathInSite  string   `json:"path"`
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

// IsPlaceholder returns true if the page is not a file on disk.
func (p *Page) IsPlaceholder() bool {
	return p.PathInGraph == ""
}

// IsPublic returns true if the page root is public.
func (p *Page) IsPublic() bool {
	return p.Root.IsPublic()
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

func (p *Page) String() string {
	return fmt.Sprintf("<Page: %s>", p.Name)
}

// SetRoot assign's page root block and sets AllBlocks to root's branches.
func (p *Page) SetRoot(root *Block) {
	p.Root = root
	p.AllBlocks = []*Block{}
	p.addTree(root)
}

func (p *Page) addTree(block *Block) {
	p.AllBlocks = append(p.AllBlocks, block)
	for _, child := range block.Children {
		p.addTree(child)
	}
}
