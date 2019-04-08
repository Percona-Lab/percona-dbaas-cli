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
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Config struct {
	ClusterName string
	PXC         PodSpec
	Proxy       PodSpec
}

type PodSpec struct {
	Size     int
	Storage  string
	Requests Resources
}

type Resources struct {
	Memory string
	CPU    string
}

// Set sets configuration parameters from given cli flags
// TODO: pre-parse resources with k8s/resource.ParseQuantity in order to return error before creating a cluster
func (c *Config) Set(f *pflag.FlagSet) (err error) {
	c.PXC.Storage, err = f.GetString("storage")
	if err != nil {
		return errors.New("undefined `storage`")
	}
	c.PXC.Size, err = f.GetInt("pxc-instances")
	if err != nil {
		return errors.New("undefined `pxc-instances`")
	}
	c.PXC.Requests.CPU, err = f.GetString("pxc-request-cpu")
	if err != nil {
		return errors.New("undefined `pxc-request-cpurage`")
	}
	c.PXC.Requests.Memory, err = f.GetString("pxc-request-mem")
	if err != nil {
		return errors.New("undefined `pxc-request-mem`")
	}

	c.Proxy.Storage, err = f.GetString("proxy-storage")
	if err != nil {
		return errors.New("undefined `storage`")
	}
	c.Proxy.Size, err = f.GetInt("proxy-instances")
	if err != nil {
		return errors.New("undefined `proxy-instances`")
	}
	c.Proxy.Requests.CPU, err = f.GetString("proxy-request-cpu")
	if err != nil {
		return errors.New("undefined `proxy-request-cpurage`")
	}
	c.Proxy.Requests.Memory, err = f.GetString("proxy-request-mem")
	if err != nil {
		return errors.New("undefined `proxy-request-mem`")
	}

	return nil
}
