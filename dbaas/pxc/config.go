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
	"k8s.io/apimachinery/pkg/api/resource"
)

type Config struct {
	ClusterName string
	PXC         PodSpec
	Proxy       PodSpec
}

type PodSpec struct {
	Size                int
	StorageSize         string
	StorageClassName    string
	Requests            Resources
	AffinityTopologyKey string
}

type AffinityTopologyKey string

var affinityValidTopologyKeys = map[string]struct{}{
	"none":                                     struct{}{},
	"kubernetes.io/hostname":                   struct{}{},
	"failure-domain.beta.kubernetes.io/zone":   struct{}{},
	"failure-domain.beta.kubernetes.io/region": struct{}{},
}

type Resources struct {
	Memory string
	CPU    string
}

// Set sets configuration parameters from given cli flags
func (c *Config) Set(f *pflag.FlagSet) (err error) {
	c.PXC.StorageSize, err = f.GetString("storage-size")
	if err != nil {
		return errors.New("undefined `storage-size`")
	}
	_, err = resource.ParseQuantity(c.PXC.StorageSize)
	if err != nil {
		return errors.Wrap(err, "storage-size")
	}

	c.PXC.StorageClassName, err = f.GetString("storage-class")
	if err != nil {
		return errors.New("undefined `storage-class`")
	}
	if c.PXC.StorageClassName != "" {
		c.PXC.StorageClassName = `storageClassName: "` + c.PXC.StorageClassName + `"`
	}

	c.PXC.Size, err = f.GetInt("pxc-instances")
	if err != nil {
		return errors.New("undefined `pxc-instances`")
	}
	c.PXC.Requests.CPU, err = f.GetString("pxc-request-cpu")
	if err != nil {
		return errors.New("undefined `pxc-request-cpu`")
	}
	_, err = resource.ParseQuantity(c.PXC.Requests.CPU)
	if err != nil {
		return errors.Wrap(err, "pxc-request-cpu")
	}

	c.PXC.Requests.Memory, err = f.GetString("pxc-request-mem")
	if err != nil {
		return errors.New("undefined `pxc-request-mem`")
	}
	_, err = resource.ParseQuantity(c.PXC.Requests.Memory)
	if err != nil {
		return errors.Wrap(err, "pxc-request-mem")
	}
	c.PXC.AffinityTopologyKey, err = f.GetString("pxc-anti-affinity-key")
	if err != nil {
		return errors.New("undefined `pxc-anti-affinity-key`")
	}
	if _, ok := affinityValidTopologyKeys[c.PXC.AffinityTopologyKey]; !ok {
		return errors.Errorf("invalid `pxc-anti-affinity-key` value: %s", c.PXC.AffinityTopologyKey)
	}

	c.Proxy.Size, err = f.GetInt("proxy-instances")
	if err != nil {
		return errors.New("undefined `proxy-instances`")
	}
	c.Proxy.Requests.CPU, err = f.GetString("proxy-request-cpu")
	if err != nil {
		return errors.New("undefined `proxy-request-cpu`")
	}
	_, err = resource.ParseQuantity(c.Proxy.Requests.CPU)
	if err != nil {
		return errors.Wrap(err, "proxy-request-cpu")
	}

	c.Proxy.Requests.Memory, err = f.GetString("proxy-request-mem")
	if err != nil {
		return errors.New("undefined `proxy-request-Memory`")
	}
	_, err = resource.ParseQuantity(c.Proxy.Requests.Memory)
	if err != nil {
		return errors.Wrap(err, "proxy-request-mem")
	}
	c.Proxy.AffinityTopologyKey, err = f.GetString("proxy-anti-affinity-key")
	if err != nil {
		return errors.New("undefined `proxy-anti-affinity-key`")
	}
	if _, ok := affinityValidTopologyKeys[c.Proxy.AffinityTopologyKey]; !ok {
		return errors.Errorf("invalid `proxy-anti-affinity-key` value: %s", c.Proxy.AffinityTopologyKey)
	}

	return nil
}
