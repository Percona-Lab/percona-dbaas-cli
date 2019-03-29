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
	"bytes"
	"encoding/base64"

	"github.com/pkg/errors"
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
	pxc := &PXC{
		name:   name,
		obj:    objects[version],
		config: Config{ClusterName: name},
	}

	return pxc, nil
}

func (p PXC) Bundle() string {
	return p.obj.Bundle
}

func (p PXC) Secrets() (string, error) {
	pass := dbaas.GenSecrets(p.obj.Secrets.Keys)
	pb64 := make(map[string]string, len(pass)+1)

	pb64["ClusterName"] = p.name

	for k, v := range pass {
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
		base64.StdEncoding.Encode(buf, v)
		pb64[k] = string(buf)
	}

	var buf bytes.Buffer
	err := p.obj.Secrets.Data.Execute(&buf, pb64)
	if err != nil {
		return "", errors.Wrap(err, "parse template:")
	}

	return buf.String(), nil
}

func (p PXC) App() (string, error) {
	var buf bytes.Buffer
	err := p.obj.CR.Execute(&buf, p.config)
	if err != nil {
		return "", errors.Wrap(err, "parse template:")
	}
	return buf.String(), nil
}

func (p *PXC) SetConfig(f *pflag.FlagSet) error {
	return p.config.Set(f)
}
