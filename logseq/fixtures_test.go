package logseq_test

import (
	"github.com/brianvoe/gofakeit/v7"

	"export-logseq/logseq"
)

func ExternalResource() logseq.ExternalResource {
	return logseq.ExternalResource{
		Uri: gofakeit.URL(),
	}
}

func LinkLabel() string {
	return gofakeit.Word()
}