package main

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"export-logseq/hugo"
	"export-logseq/logseq"
)

type SelectedPages string

const (
	AllPages    SelectedPages = "all"
	PublicPages SelectedPages = "public"
)

type ExportCmd struct {
	GraphDir      string        `env:"GRAPH_DIR" arg:""            help:"Path to the Logseq graph directory."`
	SiteDir       string        `env:"SITE_DIR" arg:""            help:"Path to the site directory."`
	SelectedPages SelectedPages `enum:"all,public" default:"public" help:"Select pages to export."`
}

func (cmd *ExportCmd) Run() error {
	graph, err := logseq.LoadGraph(cmd.GraphDir)

	if err != nil {
		return errors.Wrap(err, "loading graph")
	}

	if cmd.SelectedPages == PublicPages {
		graph = graph.PublicGraph()
	}

	log.Infof("Graph has %d pages and %d assets", len(graph.Pages), len(graph.Assets))
	if err := hugo.ExportGraph(graph, cmd.SiteDir); err != nil {
		return errors.Wrap(err, "exporting graph")
	}

	return nil
}

type CLI struct {
	Export ExportCmd `cmd:"" help:"Export a Logseq graph to SSG content folder."`
}

func main() {
	log.SetLevel(log.InfoLevel)
	log.Info("Initializing...")

	var cli CLI

	start := time.Now()
	ctx := kong.Parse(&cli)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)

	elapsed := time.Since(start)

	log.Info("All done!")
	log.Infof("Elapsed time: %s", elapsed)
}
