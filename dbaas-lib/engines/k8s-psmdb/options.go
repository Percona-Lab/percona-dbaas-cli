package psmdb

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

func (p *PSMDB) ParseOptions(opts string) error {
	err := options.Parse(&p.conf, reflect.TypeOf(p.conf), opts)
	if err != nil {
		return err
	}

	return nil
}
