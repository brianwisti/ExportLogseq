package graph_test

import (
	"github.com/brianvoe/gofakeit/v7"

	"export-logseq/graph"
)

func BlockContent() *graph.BlockContent {
	page := Page()

	return page.Root.Content
}

func Page() graph.Page {
	page := graph.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInGraph = gofakeit.Word()

	return page
}

func PageName() string {
	return gofakeit.Word()
}

func LinkLabel() string {
	return gofakeit.Word()
}
