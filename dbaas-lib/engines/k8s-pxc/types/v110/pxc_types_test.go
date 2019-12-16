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
			Image:   "test-proxy-image",
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
			Configuration: "test",
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

	if cr.Spec.ProxySQL.Enabled != c.ProxySQL.Enabled {
		t.Error("ProxySQL.Enabled !=", c.ProxySQL.Enabled)
	}
	if cr.Spec.ProxySQL.Size != c.ProxySQL.Size {
		t.Error("ProxySQL.Size !=", c.ProxySQL.Size)
	}
	if cr.Spec.ProxySQL.Image != c.ProxySQL.Image {
		t.Error("ProxySQL.Image !=", c.ProxySQL.Image)
	}
	if cr.Spec.ProxySQL.Resources.Requests.CPU != c.ProxySQL.Resources.Requests.CPU {
		t.Error("ProxySQL.Resources.Requests.CPU !=", c.ProxySQL.Resources.Requests.CPU)
	}
	if cr.Spec.ProxySQL.Resources.Requests.Memory != c.ProxySQL.Resources.Requests.Memory {
		t.Error("ProxySQL.Resources.Requests.CPU !=", c.ProxySQL.Resources.Requests.Memory)
	}
	if cr.Spec.ProxySQL.VolumeSpec.EmptyDir.Medium != c.ProxySQL.VolumeSpec.EmptyDir.Medium {
		t.Error("ProxySQL.VolumeSpec.EmptyDir.Medium !=", c.ProxySQL.VolumeSpec.EmptyDir.Medium)
	}
	if *cr.Spec.ProxySQL.Affinity.TopologyKey != c.ProxySQL.Affinity.TopologyKey {
		t.Error("ProxySQL.Affinity.TopologyKey !=", c.ProxySQL.Affinity.TopologyKey)
	}
	if cr.Spec.ProxySQL.NodeSelector["test"] != c.ProxySQL.NodeSelector["test"] {
		t.Error(`ProxySQL.NodeSelector["test"]!=`, c.ProxySQL.NodeSelector["test"])
	}
	if cr.Spec.ProxySQL.Tolerations[0].Value != c.ProxySQL.Tolerations[0].Value {
		t.Error("ProxySQL.Tolerations[0].Value !=", c.ProxySQL.Tolerations[0].Value)
	}
	if cr.Spec.ProxySQL.Tolerations[0].Key != c.ProxySQL.Tolerations[0].Key {
		t.Error("ProxySQL.Tolerations[0].Key !=", c.ProxySQL.Tolerations[0].Key)
	}
	if cr.Spec.ProxySQL.PriorityClassName != c.ProxySQL.PriorityClassName {
		t.Error("ProxySQL.PriorityClassName !=", c.ProxySQL.PriorityClassName)
	}
	if cr.Spec.ProxySQL.Annotations["test"] != c.ProxySQL.Annotations["test"] {
		t.Error(`ProxySQL.Annotations["test"]!=`, c.ProxySQL.Annotations["test"])
	}
	if cr.Spec.ProxySQL.Labels["test"] != c.ProxySQL.Labels["test"] {
		t.Error(`ProxySQL.Labels["test"]!=`, c.ProxySQL.Labels["test"])
	}
	if cr.Spec.ProxySQL.ImagePullSecrets[0].Name != c.ProxySQL.ImagePullSecrets[0].Name {
		t.Error("ProxySQL.ImagePullSecrets[0].Name !=", c.ProxySQL.ImagePullSecrets[0].Name)
	}
	if cr.Spec.ProxySQL.AllowUnsafeConfig != c.ProxySQL.AllowUnsafeConfig {
		t.Error("ProxySQL.AllowUnsafeConfig !=", c.ProxySQL.AllowUnsafeConfig)
	}
	if cr.Spec.ProxySQL.Configuration != c.ProxySQL.Configuration {
		t.Error("ProxySQL.Configurationg !=", c.ProxySQL.Configuration)
	}
	if cr.Spec.ProxySQL.PodDisruptionBudget.MinAvailable.StrVal != c.ProxySQL.PodDisruptionBudget.MinAvailable.StrVal {
		t.Error("ProxySQL.PodDisruptionBudget.MinAvailable.StrVal !=", c.ProxySQL.PodDisruptionBudget.MinAvailable.StrVal)
	}
}
