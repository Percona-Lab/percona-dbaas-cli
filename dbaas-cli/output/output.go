package output

import (
	"bytes"
	"fmt"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/pb"
	log "github.com/sirupsen/logrus"
)

func GetFormatter(format string) log.Formatter {
	switch format {
	case "json":
		return &log.JSONFormatter{
			DisableTimestamp: true,
			PrettyPrint:      true,
		}

	default:
		return &cliTextFormatter{log.TextFormatter{}}
	}
}

func GetDotprinter(format string) pb.ProgressBar {
	switch format {
	case "json":
		return pb.NewNoOp()
	default:
		return pb.NewDotPrinter()
	}
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
