package pxc

import "github.com/pkg/errors"

func (p PXC) Delete(delPVC bool, ok chan<- string, errc chan<- error) {
	err := p.Cmd.DeleteObject(p.typ, p.name)
	if err != nil {
		errc <- errors.Wrap(err, "delete cluster")
		return
	}
	if delPVC {
		err := p.Cmd.DeletePVC(p.OperatorName(), p.name)
		if err != nil {
			errc <- errors.Wrap(err, "delete cluster PVCs")
			return
		}
	}

	ok <- ""
}
