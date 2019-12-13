package v110

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

	corev1 "k8s.io/api/core/v1"
)

func TestSetNew(t *testing.T) {
	var storeMedium corev1.StorageMedium
	storeMedium = "test"
	emptyDir := corev1.EmptyDirVolumeSource{
		Medium: storeMedium,
	}
	c := config.ClusterConfig{
		PXC: config.Spec{
			Enabled: true,
			Size:    int32(3),
			Image:   "test-pxc-image",
			Resources: config.PodResources{
				Requests: config.ResourcesList{
					CPU:    "600m",
					Memory: "1G",
				},
			},
			VolumeSpec: config.VolumeSpec{
				EmptyDir: &emptyDir,
			},
			Affinity: config.PodAffinity{
				TopologyKey: "none",
			},
			NodeSelector: map[string]string{"test": "test"},
			Tolerations: []corev1.Toleration{
				{
					Key:   "testKey",
					Value: "testValue",
				},
			},
			PriorityClassName: "test",
			Annotations:       map[string]string{"test": "test"},
			Labels:            map[string]string{"test": "test"},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "test",
				},
			},
			AllowUnsafeConfig: true,
			Configuration:     "test",
			PodDisruptionBudget: config.PodDisruptionBudgetSpec{
				MinAvailable: intstr.IntOrString{
					StrVal: "test",
				},
			},
		},
		ProxySQL: config.Spec{
			Enabled: true,
			Size:    int32(3),
			Image:   "test-pxc-image",
			Resources: config.PodResources{
				Requests: config.ResourcesList{
					CPU:    "600m",
					Memory: "1G",
				},
			},
			VolumeSpec: config.VolumeSpec{
				EmptyDir: &emptyDir,
			},
			Affinity: config.PodAffinity{
				TopologyKey: "none",
			},
			NodeSelector: map[string]string{"test": "test"},
			Tolerations: []corev1.Toleration{
				{
					Key:   "testKey",
					Value: "testValue",
				},
			},
			PriorityClassName: "test",
			Annotations:       map[string]string{"test": "test"},
			Labels:            map[string]string{"test": "test"},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "test",
				},
			},
			AllowUnsafeConfig: true,
			Configuration:     "test",
			PodDisruptionBudget: config.PodDisruptionBudgetSpec{
				MinAvailable: intstr.IntOrString{
					StrVal: "test",
				},
			},
		},
	}
	cr := PerconaXtraDBCluster{}
	var s3stor *k8s.BackupStorageSpec
	cr.SetNew(c, s3stor, k8s.PlatformKubernetes)
	//if cr.Spec.PXC.Enabled != c.PXC.Enabled {
	//	t.Error("PXC.Enabled !=", c.PXC.Enabled)
	//}

	if cr.Spec.PXC.Size != c.PXC.Size {
		t.Error("PXC.Size !=", c.PXC.Size)
	}
	if cr.Spec.PXC.Image != c.PXC.Image {
		t.Error("PXC.Image !=", c.PXC.Image)
	}
	if cr.Spec.PXC.Resources.Requests.CPU != c.PXC.Resources.Requests.CPU {
		t.Error("PXC.Resources.Requests.CPU !=", c.PXC.Resources.Requests.CPU)
	}
	if cr.Spec.PXC.Resources.Requests.Memory != c.PXC.Resources.Requests.Memory {
		t.Error("PXC.Resources.Requests.CPU !=", c.PXC.Resources.Requests.Memory)
	}
	if cr.Spec.PXC.VolumeSpec.EmptyDir.Medium != c.PXC.VolumeSpec.EmptyDir.Medium {
		t.Error("PXC.VolumeSpec.EmptyDir.Medium !=", c.PXC.VolumeSpec.EmptyDir.Medium)
	}
	if *cr.Spec.PXC.Affinity.TopologyKey != c.PXC.Affinity.TopologyKey {
		t.Error("PXC.Affinity.TopologyKey !=", c.PXC.Affinity.TopologyKey)
	}
	if cr.Spec.PXC.NodeSelector["test"] != c.PXC.NodeSelector["test"] {
		t.Error(`PXC.NodeSelector["test"]!=`, c.PXC.NodeSelector["test"])
	}
	if cr.Spec.PXC.Tolerations[0].Value != c.PXC.Tolerations[0].Value {
		t.Error("PXC.Tolerations[0].Value !=", c.PXC.Tolerations[0].Value)
	}
	if cr.Spec.PXC.Tolerations[0].Key != c.PXC.Tolerations[0].Key {
		t.Error("PXC.Tolerations[0].Key !=", c.PXC.Tolerations[0].Key)
	}
	if cr.Spec.PXC.PriorityClassName != c.PXC.PriorityClassName {
		t.Error("PXC.PriorityClassName !=", c.PXC.PriorityClassName)
	}
	if cr.Spec.PXC.Annotations["test"] != c.PXC.Annotations["test"] {
		t.Error(`PXC.Annotations["test"]!=`, c.PXC.Annotations["test"])
	}
	if cr.Spec.PXC.Labels["test"] != c.PXC.Labels["test"] {
		t.Error(`PXC.Labels["test"]!=`, c.PXC.Labels["test"])
	}
	if cr.Spec.PXC.ImagePullSecrets[0].Name != c.PXC.ImagePullSecrets[0].Name {
		t.Error("PXC.ImagePullSecrets[0].Name !=", c.PXC.ImagePullSecrets[0].Name)
	}
	if cr.Spec.PXC.AllowUnsafeConfig != c.PXC.AllowUnsafeConfig {
		t.Error("PXC.AllowUnsafeConfig !=", c.PXC.AllowUnsafeConfig)
	}
	if cr.Spec.PXC.Configuration != c.PXC.Configuration {
		t.Error("PXC.Configurationg !=", c.PXC.Configuration)
	}
	if cr.Spec.PXC.PodDisruptionBudget.MinAvailable.StrVal != c.PXC.PodDisruptionBudget.MinAvailable.StrVal {
		t.Error("PXC.PodDisruptionBudget.MinAvailable.StrVal !=", c.PXC.PodDisruptionBudget.MinAvailable.StrVal)
	}
}
