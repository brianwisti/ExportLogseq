package logseq

import (
	"fmt"
	"os"

	log "github.com/charmbracelet/log"
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

	return f, nil
}
