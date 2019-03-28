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
	"github.com/spf13/pflag"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Version string

var (
	Version030 Version = "0.3.0"
)

type PXC struct {
	name   string
	config Config
	obj    dbaas.Objects
}

func New(name string, version Version) (*PXC, error) {
	return &PXC{name: name}, nil
}

func (p PXC) Bundle() string {
	return p.obj.Bundle
}

func (p PXC) Secrets() string {
	return p.obj.Secrets
}

func (p PXC) App(config map[string]string) string {
	return p.obj.CR
}

func (p PXC) SetConfig(f *pflag.FlagSet) error {
	return p.config.Set(f)
}
