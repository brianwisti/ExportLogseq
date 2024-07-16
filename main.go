package main

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"export-logseq/hugo"
	"export-logseq/logseq"
)

type EnvFlag string

// BeforeResolve loads .env file before resolving the command line arguments.
func (c EnvFlag) BeforeReset(ctx *kong.Context, trace *kong.Path) error {
	path := string(ctx.FlagValue(trace.Flag).(EnvFlag)) //nolint
	path = kong.ExpandPath(path)
	log.Infof("Loading .env file from %s", path)

	if err := godotenv.Load(path); err != nil {
		return err
	}

	return nil
}

type SelectedPages string

const (
	AllPages    SelectedPages = "all"
	PublicPages SelectedPages = "public"
)

type ExportCmd struct {
	GraphDir      string        `arg:""           env:"GRAPH_DIR"   help:"Path to the Logseq graph directory."`
	SiteDir       string        `arg:""           env:"SITE_DIR"    help:"Path to the site directory."`
	SelectedPages SelectedPages `default:"public" enum:"all,public" help:"Select pages to export."`
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
	EnvFile EnvFlag
	Export  ExportCmd `cmd:"" help:"Export a Logseq graph to SSG content folder."`
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
