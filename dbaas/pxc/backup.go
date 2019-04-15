package pxc

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Backup struct {
	cluster  string
	valolume string
	config   *PerconaXtraDBBackup
}

func NewBackup(cluster string) *Backup {
	return &Backup{
		cluster: cluster,
		config:  &PerconaXtraDBBackup{},
	}
}

func (b *Backup) Setup(storage string) {
	b.config.SetNew(b.cluster, storage)
}

func (b *Backup) CR() (string, error) {
	cr, err := json.Marshal(b.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}
