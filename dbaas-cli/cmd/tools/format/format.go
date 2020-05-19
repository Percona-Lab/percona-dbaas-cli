package format

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Detect(cmd *cobra.Command) error {
	format, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	switch format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: true,
			PrettyPrint:      true,
		})
	default:
		log.SetFormatter(&cliTextFormatter{log.TextFormatter{}})
	}
	return nil
}

type cliTextFormatter struct {
	log.TextFormatter
}

func (f *cliTextFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	if entry.Level == log.ErrorLevel {
		b.WriteString("[Error] " + entry.Message)
	}
	if entry.Message != "" && entry.Level != log.ErrorLevel && entry.Message != "information" {
		b.WriteString(entry.Message + "\n")
	}

	if len(entry.Data) == 0 {
		b.WriteString("\n")
		return b.Bytes(), nil
	}

	for _, v := range entry.Data {
		fmt.Fprint(b, v)
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}
