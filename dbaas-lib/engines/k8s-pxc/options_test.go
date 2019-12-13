package pxc

import (
	"fmt"
	"log"
	"testing"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

	corev1 "k8s.io/api/core/v1"
)

type Case struct {
	Name    string
	Options string
}

func TestOptions(t *testing.T) {
	res := config.PodResources{
		Requests: config.ResourcesList{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	aff := config.PodAffinity{
		TopologyKey: "none",
	}

	cases := []struct {
		name    string
		options string
		desired config.ClusterConfig
	}{
		{
			name:    "pxc first level",
			options: "pxc.size=5",
			desired: config.ClusterConfig{
				PXC: config.Spec{
					Size:      5,
					Resources: res,
					Affinity:  aff,
				},
				ProxySQL: config.Spec{
					Size:      int32(1),
					Resources: res,
					Affinity:  aff,
				},
				S3: k8s.S3StorageConfig{
					SkipStorage: true,
				},
			},
		},
		{
			name:    "pxc resources",
			options: "pxc.resources.requests.cpu=300m,pxc.resources.requests.memory=0.5G",
			desired: config.ClusterConfig{
				PXC: config.Spec{
					Size: int32(3),
					Resources: config.PodResources{
						Requests: config.ResourcesList{
							CPU:    "300m",
							Memory: "0.5G",
						},
					},
					Affinity: aff,
				},
				ProxySQL: config.Spec{
					Size:      int32(1),
					Resources: res,
					Affinity:  aff,
				},
				S3: k8s.S3StorageConfig{
					SkipStorage: true,
				},
			},
		},
		{
			name:    "pxc volume spec",
			options: "pxc.volumeSpec.hostPath.path=test",
			desired: config.ClusterConfig{
				PXC: config.Spec{
					Size:      int32(3),
					Resources: res,
					Affinity:  aff,
					VolumeSpec: config.VolumeSpec{
						HostPath: corev1.HostPathVolumeSource{
							Path: "test",
						},
					},
				},
				ProxySQL: config.Spec{
					Size:      int32(1),
					Resources: res,
					Affinity:  aff,
				},
				S3: k8s.S3StorageConfig{
					SkipStorage: true,
				},
			},
		},
	}

	for _, c := range cases {
		pxc := PXC{}
		err := pxc.ParseOptions(c.options)
		if err != nil {
			t.Error(c.name, err)
		}
		if fmt.Sprintln(c.desired) != fmt.Sprintln(pxc.config) {
			t.Error(c.name, "wrong result")
		}
		log.Println(c.name, "- ok!")
	}
}
