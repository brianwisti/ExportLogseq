package logseq

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"olympos.io/encoding/edn"
)

// Describes the aspects of Logseq configuration we are interested in.
type LogseqConfig struct {
	FileNameFormat  edn.Keyword `edn:"file/name-format"`
	PreferredFormat edn.Keyword `edn:"preferred-format"`
}

func CheckConfig(configFile string) error {
	log.Info("Loading config file:", configFile)

	logseqConfigBytes, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "reading config file")
	}

	var f LogseqConfig
	ednErr := edn.Unmarshal(logseqConfigBytes, &f)

	if ednErr != nil {
		return errors.Wrap(ednErr, "unmarshalling EDN")
	}

	// We're specifically catering to my graph first.
	if f.FileNameFormat != "triple-lowbar" {
		return errors.New("unsupported file name format" + f.FileNameFormat.String())
	}

	if f.PreferredFormat != "markdown" {
		return errors.New("unsupported preferred format" + f.PreferredFormat.String())
	}

	return nil
}
