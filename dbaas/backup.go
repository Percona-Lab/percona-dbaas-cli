package dbaas

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Backupper interface {
	CR() (string, error)

	Name() string
	OperatorName() string

	CheckStatus(data []byte) (ClusterState, []string, error)
	CheckOperatorLogs(data []byte) ([]OutuputMsg, error)
}

type BackupState string

const (
	BackupUnknown   BackupState = "Unknown"
	BackupStarting              = "Starting"
	BackupRunning               = "Running"
	BackupFailed                = "Failed"
	BackupSucceeded             = "Succeeded"
)

func Backup(typ string, app Backupper, ok chan<- string, msg chan<- OutuputMsg, errc chan<- error) {
	cr, err := app.CR()
	if err != nil {
		errc <- errors.Wrap(err, "create backup cr")
		return
	}

	err = apply(cr)
	if err != nil {
		errc <- errors.Wrap(err, "apply backup cr")
		return
	}
	time.Sleep(1 * time.Minute)

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := getCR(typ, app.Name())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := app.CheckStatus(status)
		if err != nil {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}

		switch state {
		case ClusterStateReady:
			ok <- strings.Join(msgs, "\n")
			return
		case ClusterStateError:
			errc <- errors.New(strings.Join(msgs, "\n"))
			return
		case ClusterStateInit:
		}

		opLogsStream, err := readOperatorLogs(app.OperatorName())
		if err != nil {
			errc <- errors.Wrap(err, "get operator logs")
			return
		}

		opLogs, err := app.CheckOperatorLogs(opLogsStream)
		if err != nil {
			errc <- errors.Wrap(err, "parse operator logs")
			return
		}

		for _, entry := range opLogs {
			msg <- entry
		}

		if tries >= getStatusMaxTries {
			errc <- errors.Wrap(err, "unable to start cluster")
			return
		}

		tries++
	}
}
