package pxc

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

// ParseOptions parse PXC options given in "object.paramValue=val,objectTwo.paramValue=val" string
func (p *PXC) ParseOptions(opts string) error {
	err := options.Parse(&p.conf, reflect.TypeOf(p.conf), opts)
	if err != nil {
		return err
	}

	return nil
}
