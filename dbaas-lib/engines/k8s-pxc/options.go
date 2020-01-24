package pxc

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

// ParseOptions parse PXC options given in "object.paramValue=val,objectTwo.paramValue=val" string
func (p *PXC) ParseOptions(opts string) error {
	c := objects[defaultVersion].pxc
	c.SetDefaults()
	err := options.Parse(&c, reflect.TypeOf(c), opts)
	if err != nil {
		return err
	}
	p.conf = c

	return nil
}
