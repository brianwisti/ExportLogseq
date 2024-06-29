package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lpernett/godotenv"
	log "github.com/sirupsen/logrus"
	"olympos.io/encoding/edn"
)

// Describes the aspects of Logseq configuration we are interested in
type LogseqConfig struct {
	FileNameFormat  edn.Keyword `edn:"file/name-format"`
	PreferredFormat edn.Keyword `edn:"preferred-format"`
}

func LoadConfig(configFile string) (LogseqConfig, error) {
	log.Info("Loading config file:", configFile)
	logseqConfigBytes, err := os.ReadFile(configFile)
	if err != nil {
		return LogseqConfig{}, fmt.Errorf("reading config file: %v", err)
	}

	var f LogseqConfig
	ednErr := edn.Unmarshal(logseqConfigBytes, &f)
	if ednErr != nil {
		log.Fatalf("Error unmarshalling EDN: %v", ednErr)
		return LogseqConfig{}, fmt.Errorf("unmarshalling EDN: %v", ednErr)
	}

	log.Info("Unmarshalled EDN", f)

	if f.FileNameFormat != "triple-lowbar" {
		return f, fmt.Errorf("unsupported file/name-format: %v", f.FileNameFormat)
	}

	if f.PreferredFormat != "markdown" {
		return f, fmt.Errorf("unsupported preferred-format: %v", f.PreferredFormat)
	}

	return f, nil
}

func main() {
	log.Info("Initializing...")
	godotenv.Load()
	graphDir := os.Getenv("GRAPH_DIR")
	log.Info("GRAPH_DIR:", graphDir)
	if graphDir == "" {
		log.Fatal("GRAPH_DIR is not set in .env file or environment variables")
	}
	configFile := filepath.Join(graphDir, "logseq", "config.edn")

	logseqConfigBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	var f LogseqConfig
	ednErr := edn.Unmarshal(logseqConfigBytes, &f)
	if ednErr != nil {
		log.Fatalf("Error unmarshalling EDN: %v", ednErr)
	}
	log.Info("Unmarshalled EDN", f)
	log.Info("File Name Format:", f.FileNameFormat)

	// We're specifically catering to my graph first.
	if f.FileNameFormat != "triple-lowbar" {
		log.Fatal("Unsupported file name format:", f.FileNameFormat)
	}

	if f.PreferredFormat != "markdown" {
		log.Fatal("Unsupported preferred format:", f.PreferredFormat)
	}

	log.Info("All done!")
}
