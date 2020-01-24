package psmdb

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

func (p *PSMDB) ParseOptions(opts string) error {
	c := objects[defaultVersion].psmdb
	c.SetDefaults()
	err := options.Parse(&c, reflect.TypeOf(c), opts)
	if err != nil {
		return err
	}
	p.conf = c

	return nil
}
