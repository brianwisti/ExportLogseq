package logseq_test

import (
	"github.com/brianvoe/gofakeit/v7"

	"export-logseq/logseq"
)

func BlockContent() *logseq.BlockContent {
	page := Page()
	return page.Root.Content
}

func Page() logseq.Page {
	page := logseq.NewEmptyPage()
	page.Name = gofakeit.Word()
	page.PathInSite = gofakeit.Word()
	page.PathInGraph = gofakeit.Word()

	return page
}

func PageName() string {
	return gofakeit.Word()
}

func LinkLabel() string {
	return gofakeit.Word()
}
