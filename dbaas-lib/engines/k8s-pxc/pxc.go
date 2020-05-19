// Copyright © 2019 Percona, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pxc

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	v110 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v110"
	v120 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v120"
	v130 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v130"
	v140 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v140"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
)

const (
	provider               = "k8s"
	engine                 = "pxc"
	defaultVersion Version = "1.4.0"
)

type Version string

var objects map[Version]VersionObject

func init() {
	// Register pxc engine in dbaas
	pxc, err := NewPXCController("", "k8s")
	if err != nil {
		fmt.Println("Cant start. Setup your kubectl")
		os.Exit(1)
	}
	dbaas.RegisterEngine(provider, engine, pxc)

	// Register pxc versions
	objects = make(map[Version]VersionObject)
	objects["1.1.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v110.Bundle,
		},
		pxc: &v110.PerconaXtraDBCluster{},
	}
	objects["1.2.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v120.Bundle,
		},
		pxc: &v120.PerconaXtraDBCluster{},
	}
	objects["1.3.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v130.Bundle,
		},
		pxc: &v130.PerconaXtraDBCluster{},
	}
	objects["1.4.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v140.Bundle,
		},
		pxc: &v140.PerconaXtraDBCluster{},
	}
}

// PXC represents PXC Operator controller
type PXC struct {
	cmd          *k8s.Cmd
	conf         PXDBCluster
	platformType k8s.PlatformType
	bundle       []k8s.BundleObject
}

type VersionObject struct {
	k8s k8s.Objects
	pxc PXDBCluster
}

// NewPXCController returns new PXCOperator Controller
func NewPXCController(envCrt, provider string) (*PXC, error) {
	var pxc PXC
	if len(provider) == 0 || provider == "k8s" {
		k8sCmd, err := k8s.New(envCrt)
		if err != nil {
			return nil, errors.Wrap(err, "new Cmd")
		}
		pxc.cmd = k8sCmd
		pxc.platformType = k8sCmd.GetPlatformType()
	}

	return &pxc, nil
}

func (p *PXC) setVersionObjectsWithDefaults(version Version) error {
	if p.conf != nil && p.bundle != nil {
		return nil
	}
	switch i := len(version); {
	case i == 0:
		version = defaultVersion
	default:
		if _, ok := objects[version]; !ok {
			return errors.Errorf("unsupporeted version %s", version)
		}
	}

	p.conf = objects[version].pxc
	err := p.conf.SetDefaults()
	if err != nil {
		errors.Wrap(err, "set defaults")
	}
	p.bundle = objects[version].k8s.Bundle

	return nil
}

func (p PXC) getCR(cluster PXDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PXC) operatorName() string {
	return "percona-xtradb-cluster-operator"
}
