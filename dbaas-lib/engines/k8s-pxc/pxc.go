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
	"strings"

	"github.com/pkg/errors"

	v110 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v110"
	v120 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v120"
	v130 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/v130"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/pdl"
)

const (
	provider               = "k8s"
	engine                 = "pxc"
	defaultVersion Version = "1.3.0"
)

var objects map[Version]VersionObject

func init() {
	// Register pxc engine in dbaas
	pxc, err := NewPXCController("", "k8s")
	if err != nil {
		fmt.Println("Cant start. Setup your kubectl")
		os.Exit(1)
	}
	pdl.RegisterEngine(provider, engine, pxc)

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
}

// PXC represents PXC Operator controller
type PXC struct {
	cmd          *k8s.Cmd
	conf         PXDBCluster
	platformType k8s.PlatformType
}

type Version string

type PXCMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type AppState string

const (
	AppStateUnknown AppState = "unknown"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

type PerconaXtraDBClusterStatus struct {
	PXC      AppStatus `json:"pxc,omitempty"`
	ProxySQL AppStatus `json:"proxysql,omitempty"`
	Host     string    `json:"host,omitempty"`
	Messages []string  `json:"message,omitempty"`
	Status   AppState  `json:"state,omitempty"`
}

type AppStatus struct {
	Size    int32    `json:"size,omitempty"`
	Ready   int32    `json:"ready,omitempty"`
	Status  AppState `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
}

type PXCResource struct {
	Meta   PXCMeta `json:"metadata"`
	Status PerconaXtraDBClusterStatus
}
type k8sCluster struct {
	Items []PXCResource `json:"items"`
}

type k8sStatus struct {
	Status PerconaXtraDBClusterStatus
}

type PVCMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	SelfLink  string `json:"selflink"`
	UID       string `json:"uid"`
}

type k8sPVC struct {
	Meta PVCMeta `json:"metadata"`
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

func (p PXC) bundle(v map[Version]VersionObject, operatorVersion string) []k8s.BundleObject {
	if operatorVersion == "" {
		operatorVersion = v[defaultVersion].pxc.GetOperatorImage()
	}

	for i, o := range v[defaultVersion].k8s.Bundle {
		if o.Kind == "Deployment" && o.Name == p.operatorName() {
			v[defaultVersion].k8s.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
	return v[defaultVersion].k8s.Bundle
}

func (p PXC) getCR(cluster PXDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PXC) operatorName() string {
	return "percona-xtradb-cluster-operator"
}
