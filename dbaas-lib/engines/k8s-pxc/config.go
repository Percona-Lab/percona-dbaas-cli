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

import "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"

// PXDBCluster represent interface for ckuster types
type PXDBCluster interface {
	Upgrade(imgs map[string]string)
	GetName() string
	SetDefaults() error
	MarshalRequests() error
	GetCR() (string, error)
	SetLabels(labels map[string]string)
	SetName(name string)
	SetUsersSecretName(name string)
	GetOperatorImage() string
	SetupMiniConfig() //For Minikube and Minishift
	GetProxysqlServiceType() string
	GetStatus() dbaas.State
	GetPXCStatus() string
	GetStatusHost() string
}
