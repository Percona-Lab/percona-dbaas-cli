package psmdb

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

func (p *PSMDB) ParseOptions(opts string) error {
	var c config.ClusterConfig

	err := options.Parse(&c, reflect.TypeOf(c), opts)
	if err != nil {
		return err
	}
	p.config = c

	return nil
}
