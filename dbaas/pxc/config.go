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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sversion "k8s.io/apimachinery/pkg/version"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

func ParseCreateFlagsToConfig(f *pflag.FlagSet) (config dbaas.ClusterConfig, err error) {
	config.PXC.StorageSize, err = f.GetString("storage-size")
	if err != nil {
		return config, errors.New("undefined `storage size`")
	}
	config.PXC.StorageClass, err = f.GetString("storage-class")
	if err != nil {
		return config, errors.New("undefined `storage class`")
	}
	config.PXC.Instances, err = f.GetInt32("pxc-instances")
	if err != nil {
		return config, errors.New("undefined `pxc-instances`")
	}
	config.PXC.RequestCPU, err = f.GetString("pxc-request-cpu")
	if err != nil {
		return config, errors.New("undefined `pxc-request-cpu`")
	}
	config.PXC.RequestMem, err = f.GetString("pxc-request-mem")
	if err != nil {
		return config, errors.New("undefined `pxc-request-mem`")
	}
	config.PXC.AntiAffinityKey, err = f.GetString("pxc-anti-affinity-key")
	if err != nil {
		return config, errors.New("undefined `pxc-anti-affinity-key`")
	}

	config.ProxySQL.Instances, err = f.GetInt32("proxy-instances")
	if err != nil {
		return config, errors.New("undefined `proxy-instances`")
	}
	if config.ProxySQL.Instances > 0 {
		config.ProxySQL.RequestCPU, err = f.GetString("proxy-request-cpu")
		if err != nil {
			return config, errors.New("undefined `proxy-request-cpu`")
		}
		config.ProxySQL.RequestMem, err = f.GetString("proxy-request-mem")
		if err != nil {
			return config, errors.New("undefined `proxy-request-mem`")
		}
		config.ProxySQL.AntiAffinityKey, err = f.GetString("proxy-anti-affinity-key")
		if err != nil {
			return config, errors.New("undefined `proxy-anti-affinity-key`")
		}
	}

	skipS3Storage, err := f.GetBool("s3-skip-storage")
	if err != nil {
		return config, errors.New("undefined `s3-skip-storage`")
	}

	if !skipS3Storage {
		config.S3.EndpointURL, err = f.GetString("s3-endpoint-url")
		if err != nil {
			return config, errors.New("undefined `s3-endpoint-url`")
		}
		config.S3.Bucket, err = f.GetString("s3-bucket")
		if err != nil {
			return config, errors.New("undefined `s3-bucket`")
		}
		config.S3.Region, err = f.GetString("s3-region")
		if err != nil {
			return config, errors.New("undefined `s3-region`")
		}
		config.S3.CredentialsSecret, err = f.GetString("s3-credentials-secret")
		if err != nil {
			return config, errors.New("undefined `s3-credentials-secret`")
		}
		config.S3.KeyID, err = f.GetString("s3-key-id")
		if err != nil {
			return config, errors.New("undefined `s3-key-id`")
		}
		config.S3.Key, err = f.GetString("s3-key")
		if err != nil {
			return config, errors.New("undefined `s3-key`")
		}
	}

	return
}

func ParseEditFlagsToConfig(f *pflag.FlagSet) (config dbaas.ClusterConfig, err error) {
	config.PXC.Instances, err = f.GetInt32("pxc-instances")
	if err != nil {
		return config, errors.New("undefined `pxc-instances`")
	}

	config.ProxySQL.Instances, err = f.GetInt32("proxy-instances")
	if err != nil {
		return config, errors.New("undefined `proxy-instances`")
	}

	return
}

func ParseAddStorageFlagsToConfig(f *pflag.FlagSet) (config dbaas.ClusterConfig, err error) {
	config.PXC.Instances, err = f.GetInt32("pxc-instances")
	if err != nil {
		return config, errors.New("undefined `pxc-instances`")
	}

	config.ProxySQL.Instances, err = f.GetInt32("proxy-instances")
	if err != nil {
		return config, errors.New("undefined `proxy-instances`")
	}

	config.S3.EndpointURL, err = f.GetString("s3-endpoint-url")
	if err != nil {
		return config, errors.New("undefined `s3-endpoint-url`")
	}
	config.S3.Bucket, err = f.GetString("s3-bucket")
	if err != nil {
		return config, errors.New("undefined `s3-bucket`")
	}
	config.S3.Region, err = f.GetString("s3-region")
	if err != nil {
		return config, errors.New("undefined `s3-region`")
	}
	config.S3.CredentialsSecret, err = f.GetString("s3-credentials-secret")
	if err != nil {
		return config, errors.New("undefined `s3-credentials-secret`")
	}
	config.S3.KeyID, err = f.GetString("s3-key-id")
	if err != nil {
		return config, errors.New("undefined `s3-key-id`")
	}
	config.S3.Key, err = f.GetString("s3-key")
	if err != nil {
		return config, errors.New("undefined `s3-key`")
	}

	return
}

// PerconaXtraDBClusterSpec defines the desired state of PerconaXtraDBCluster
type PerconaXtraDBClusterSpec struct {
	Platform    *Platform           `json:"platform,omitempty"`
	SecretsName string              `json:"secretsName,omitempty"`
	PXC         *PodSpec            `json:"pxc,omitempty"`
	ProxySQL    *PodSpec            `json:"proxysql,omitempty"`
	PMM         *PMMSpec            `json:"pmm,omitempty"`
	Backup      *PXCScheduledBackup `json:"backup,omitempty"`
}

type PXCScheduledBackup struct {
	Image            string                              `json:"image,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference       `json:"imagePullSecrets,omitempty"`
	Schedule         []PXCScheduledBackupSchedule        `json:"schedule,omitempty"`
	Storages         map[string]*dbaas.BackupStorageSpec `json:"storages,omitempty"`
}

type PXCScheduledBackupSchedule struct {
	Name        string `json:"name,omitempty"`
	Schedule    string `json:"schedule,omitempty"`
	Keep        int    `json:"keep,omitempty"`
	StorageName string `json:"storageName,omitempty"`
}
type AppState string

const (
	AppStateUnknown AppState = "unknown"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

// PerconaXtraDBClusterStatus defines the observed state of PerconaXtraDBCluster
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

// PerconaXtraDBCluster is the Schema for the perconaxtradbclusters API
type PerconaXtraDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerconaXtraDBClusterSpec   `json:"spec,omitempty"`
	Status PerconaXtraDBClusterStatus `json:"status,omitempty"`
}

// PerconaXtraDBClusterList contains a list of PerconaXtraDBCluster
type PerconaXtraDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerconaXtraDBCluster `json:"items"`
}

type PodSpec struct {
	Enabled             bool                          `json:"enabled,omitempty"`
	Size                int32                         `json:"size,omitempty"`
	Image               string                        `json:"image,omitempty"`
	Resources           *PodResources                 `json:"resources,omitempty"`
	VolumeSpec          *VolumeSpec                   `json:"volumeSpec,omitempty"`
	Affinity            *PodAffinity                  `json:"affinity,omitempty"`
	NodeSelector        map[string]string             `json:"nodeSelector,omitempty"`
	Tolerations         []corev1.Toleration           `json:"tolerations,omitempty"`
	PriorityClassName   string                        `json:"priorityClassName,omitempty"`
	Annotations         map[string]string             `json:"annotations,omitempty"`
	Labels              map[string]string             `json:"labels,omitempty"`
	ImagePullSecrets    []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	AllowUnsafeConfig   bool                          `json:"allowUnsafeConfigurations,omitempty"`
	Configuration       string                        `json:"configuration,omitempty"`
	PodDisruptionBudget *PodDisruptionBudgetSpec      `json:"podDisruptionBudget,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type PodAffinity struct {
	TopologyKey *string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    *corev1.Affinity `json:"advanced,omitempty"`
}

type PodResources struct {
	Requests *ResourcesList `json:"requests,omitempty"`
	Limits   *ResourcesList `json:"limits,omitempty"`
}

type PMMSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerHost string `json:"serverHost,omitempty"`
	Image      string `json:"image,omitempty"`
	ServerUser string `json:"serverUser,omitempty"`
}

type ResourcesList struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

type BackupStorageSpec struct {
	Type   BackupStorageType   `json:"type"`
	S3     BackupStorageS3Spec `json:"s3,omitempty"`
	Volume *VolumeSpec         `json:"volume,omitempty"`
}

type BackupStorageType string

const (
	BackupStorageFilesystem BackupStorageType = "filesystem"
	BackupStorageS3         BackupStorageType = "s3"
)

type BackupStorageS3Spec struct {
	Bucket            string `json:"bucket"`
	CredentialsSecret string `json:"credentialsSecret"`
	Region            string `json:"region,omitempty"`
	EndpointURL       string `json:"endpointUrl,omitempty"`
}

type VolumeSpec struct {
	// EmptyDir to use as data volume for mysql. EmptyDir represents a temporary
	// directory that shares a pod's lifetime.
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath to use as data volume for mysql. HostPath represents a
	// pre-existing file or directory on the host machine that is directly
	// exposed to the container.
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim to specify PVC spec for the volume for mysql data.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type Volume struct {
	PVCs    []corev1.PersistentVolumeClaim
	Volumes []corev1.Volume
}

type Platform string

const (
	PlatformUndef      Platform = ""
	PlatformKubernetes          = "kubernetes"
	PlatformOpenshift           = "openshift"
)

// ServerVersion represents info about k8s / openshift server version
type ServerVersion struct {
	Platform Platform
	Info     k8sversion.Info
}

const AffinityTopologyKeyOff = "none"

var AffinityValidTopologyKeys = map[string]struct{}{
	AffinityTopologyKeyOff:                     struct{}{},
	"kubernetes.io/hostname":                   struct{}{},
	"failure-domain.beta.kubernetes.io/zone":   struct{}{},
	"failure-domain.beta.kubernetes.io/region": struct{}{},
}

var defaultAffinityTopologyKey = "kubernetes.io/hostname"

func (cr *PerconaXtraDBCluster) UpdateWith(c dbaas.ClusterConfig, s3 *dbaas.BackupStorageSpec) (err error) {
	if _, ok := cr.Spec.Backup.Storages[dbaas.DefaultBcpStorageName]; !ok && s3 != nil {
		if cr.Spec.Backup.Storages == nil {
			cr.Spec.Backup.Storages = make(map[string]*dbaas.BackupStorageSpec)
		}

		cr.Spec.Backup.Storages[dbaas.DefaultBcpStorageName] = s3
	}

	if c.PXC.Instances > 0 {
		cr.Spec.PXC.Size = c.PXC.Instances
	}

	if c.ProxySQL.Instances > 0 {
		cr.Spec.ProxySQL.Size = c.ProxySQL.Instances
	}

	// Disable ProxySQL if size set to 0
	if cr.Spec.ProxySQL.Size == 0 {
		cr.Spec.ProxySQL.Enabled = false
	}

	return nil
}

// Upgrade upgrades culster with given images
func (cr *PerconaXtraDBCluster) Upgrade(imgs map[string]string) {
	if img, ok := imgs["pxc"]; ok {
		cr.Spec.PXC.Image = img
	}
	if img, ok := imgs["proxysql"]; ok {
		cr.Spec.ProxySQL.Image = img
	}
	if img, ok := imgs["backup"]; ok {
		cr.Spec.Backup.Image = img
	}
}

func (cr *PerconaXtraDBCluster) SetNew(clusterName string, c dbaas.ClusterConfig, s3 *dbaas.BackupStorageSpec, p dbaas.PlatformType) (err error) {
	cr.ObjectMeta.Name = clusterName
	cr.SetDefaults()

	if len(c.PXC.StorageSize) > 0 {
		volSizeFlag := c.PXC.StorageSize
		volSize, err := resource.ParseQuantity(volSizeFlag)
		if err != nil {
			return errors.Wrap(err, "storage-size")
		}
		cr.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests = corev1.ResourceList{corev1.ResourceStorage: volSize}
	}

	if len(c.PXC.StorageClass) > 0 {
		volClassNameFlag := c.PXC.StorageClass
		if volClassNameFlag != "" {
			cr.Spec.PXC.VolumeSpec.PersistentVolumeClaim.StorageClassName = &volClassNameFlag
		}
	}

	if c.PXC.Instances > 0 {
		cr.Spec.PXC.Size = c.PXC.Instances
	}

	if len(c.PXC.RequestCPU) > 0 && len(c.PXC.RequestMem) > 0 {
		pxcCPU := c.PXC.RequestCPU
		_, err = resource.ParseQuantity(pxcCPU)
		if err != nil {
			return errors.Wrap(err, "pxc-request-cpu")
		}
		pxcMemory := c.PXC.RequestMem
		_, err = resource.ParseQuantity(pxcMemory)
		if err != nil {
			return errors.Wrap(err, "pxc-request-mem")
		}
		cr.Spec.PXC.Resources = &PodResources{
			Requests: &ResourcesList{
				CPU:    pxcCPU,
				Memory: pxcMemory,
			},
		}
	}

	if len(c.PXC.AntiAffinityKey) > 0 {
		pxctpk := c.PXC.AntiAffinityKey
		if _, ok := AffinityValidTopologyKeys[pxctpk]; !ok {
			return errors.Errorf("invalid `pxc-anti-affinity-key` value: %s", pxctpk)
		}
		cr.Spec.PXC.Affinity.TopologyKey = &pxctpk
	}

	if c.ProxySQL.Instances > 0 {
		cr.Spec.ProxySQL.Size = c.ProxySQL.Instances
	}

	// Disable ProxySQL if size set to 0
	if cr.Spec.ProxySQL.Size > 0 {
		err := cr.setProxySQL(c)
		if err != nil {
			return err
		}
	} else {
		cr.Spec.ProxySQL.Enabled = false
	}

	if s3 != nil {
		cr.Spec.Backup.Storages = map[string]*dbaas.BackupStorageSpec{
			dbaas.DefaultBcpStorageName: s3,
		}
	}

	switch p {
	case dbaas.PlatformMinishift, dbaas.PlatformMinikube:
		none := AffinityTopologyKeyOff
		cr.Spec.PXC.Affinity.TopologyKey = &none
		cr.Spec.PXC.Resources = nil
		cr.Spec.ProxySQL.Affinity.TopologyKey = &none
		cr.Spec.ProxySQL.Resources = nil
	}

	return nil
}

func (cr *PerconaXtraDBCluster) setProxySQL(c dbaas.ClusterConfig) error {
	proxyCPU := c.ProxySQL.RequestCPU
	_, err := resource.ParseQuantity(proxyCPU)
	if err != nil {
		return errors.Wrap(err, "proxy-request-cpu")
	}
	proxyMemory := c.ProxySQL.RequestMem
	_, err = resource.ParseQuantity(proxyMemory)
	if err != nil {
		return errors.Wrap(err, "proxy-request-mem")
	}
	cr.Spec.ProxySQL.Resources = &PodResources{
		Requests: &ResourcesList{
			CPU:    proxyCPU,
			Memory: proxyMemory,
		},
	}

	proxytpk := c.ProxySQL.AntiAffinityKey
	if _, ok := AffinityValidTopologyKeys[proxytpk]; !ok {
		return errors.Errorf("invalid `proxy-anti-affinity-key` value: %s", proxytpk)
	}
	cr.Spec.ProxySQL.Affinity.TopologyKey = &proxytpk

	return nil
}

func (cr *PerconaXtraDBCluster) SetDefaults() {
	one := intstr.FromInt(1)

	cr.TypeMeta.APIVersion = "pxc.percona.com/v1"
	cr.TypeMeta.Kind = "PerconaXtraDBCluster"
	cr.ObjectMeta.Finalizers = []string{"delete-pxc-pods-in-order"}

	cr.Spec.SecretsName = cr.Name + "-secrets"

	cr.Spec.PXC = &PodSpec{}
	cr.Spec.PXC.Size = 3
	cr.Spec.PXC.Image = "percona/percona-xtradb-cluster-operator:1.1.0-pxc"
	cr.Spec.PXC.Affinity = &PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.PXC.PodDisruptionBudget = &PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volPXC, _ := resource.ParseQuantity("6G")
	cr.Spec.PXC.VolumeSpec = &VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volPXC},
			},
		},
	}

	cr.Spec.ProxySQL = &PodSpec{}
	cr.Spec.ProxySQL.Enabled = true
	cr.Spec.ProxySQL.Size = 1
	cr.Spec.ProxySQL.Image = "percona/percona-xtradb-cluster-operator:1.1.0-proxysql"
	cr.Spec.ProxySQL.Affinity = &PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.ProxySQL.PodDisruptionBudget = &PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volProxy, _ := resource.ParseQuantity("1G")
	cr.Spec.ProxySQL.VolumeSpec = &VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volProxy},
			},
		},
	}

	cr.Spec.Backup = &PXCScheduledBackup{
		Image: "percona/percona-xtradb-cluster-operator:1.1.0-backup",
	}
}
