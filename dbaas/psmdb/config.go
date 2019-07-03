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

package psmdb

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

// PerconaServerMongoDB is the Schema for the perconaservermongodbs API
type PerconaServerMongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerconaServerMongoDBSpec   `json:"spec,omitempty"`
	Status PerconaServerMongoDBStatus `json:"status,omitempty"`
}

// PerconaServerMongoDBList contains a list of PerconaServerMongoDB
type PerconaServerMongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerconaServerMongoDB `json:"items"`
}

type ClusterRole string

const (
	ClusterRoleShardSvr  ClusterRole = "shardsvr"
	ClusterRoleConfigSvr ClusterRole = "configsvr"
)

// PerconaServerMongoDBSpec defines the desired state of PerconaServerMongoDB
type PerconaServerMongoDBSpec struct {
	Pause            bool                          `json:"pause,omitempty"`
	Platform         *Platform                     `json:"platform,omitempty"`
	Image            string                        `json:"image,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	RunUID           int64                         `json:"runUid,omitempty"`
	UnsafeConf       bool                          `json:"allowUnsafeConfigurations"`
	Mongod           *MongodSpec                   `json:"mongod,omitempty"`
	Replsets         []*ReplsetSpec                `json:"replsets,omitempty"`
	Secrets          *SecretsSpec                  `json:"secrets,omitempty"`
	Backup           BackupSpec                    `json:"backup,omitempty"`
	ImagePullPolicy  corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	PMM              PMMSpec                       `json:"pmm,omitempty"`
}

type ReplsetMemberStatus struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type ReplsetStatus struct {
	Members     []*ReplsetMemberStatus `json:"members,omitempty"`
	ClusterRole ClusterRole            `json:"clusterRole,omitempty"`

	Initialized bool     `json:"initialized,omitempty"`
	Size        int32    `json:"size"`
	Ready       int32    `json:"ready"`
	Status      AppState `json:"status,omitempty"`
	Message     string   `json:"message,omitempty"`
}

type AppState string

const (
	AppStatePending AppState = "pending"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

// PerconaServerMongoDBStatus defines the observed state of PerconaServerMongoDB
type PerconaServerMongoDBStatus struct {
	Status     AppState                  `json:"state,omitempty"`
	Message    string                    `json:"message,omitempty"`
	Conditions []ClusterCondition        `json:"conditions,omitempty"`
	Replsets   map[string]*ReplsetStatus `json:"replsets,omitempty"`
}

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse                   = "False"
	ConditionUnknown                 = "Unknown"
)

type ClusterConditionType string

const (
	ClusterReady   ClusterConditionType = "ClusterReady"
	ClusterInit                         = "ClusterInitializing"
	ClusterRSInit                       = "ReplsetInitialized"
	ClusterRSReady                      = "ReplsetReady"
	ClusterError                        = "Error"
)

type ClusterCondition struct {
	Status             ConditionStatus      `json:"status"`
	Type               ClusterConditionType `json:"type"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

type PMMSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerHost string `json:"serverHost,omitempty"`
	Image      string `json:"image,omitempty"`
}

type MultiAZ struct {
	Affinity            *PodAffinity             `json:"affinity,omitempty"`
	NodeSelector        map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations         []corev1.Toleration      `json:"tolerations,omitempty"`
	PriorityClassName   string                   `json:"priorityClassName,omitempty"`
	Annotations         map[string]string        `json:"annotations,omitempty"`
	Labels              map[string]string        `json:"labels,omitempty"`
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type PodAffinity struct {
	TopologyKey *string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    *corev1.Affinity `json:"advanced,omitempty"`
}

type ReplsetSpec struct {
	Resources                    *ResourcesSpec `json:"resources,omitempty"`
	Name                         string         `json:"name"`
	Size                         int32          `json:"size"`
	ClusterRole                  ClusterRole    `json:"clusterRole,omitempty"`
	Arbiter                      Arbiter        `json:"arbiter,omitempty"`
	Expose                       Expose         `json:"expose,omitempty"`
	VolumeSpec                   *VolumeSpec    `json:"volumeSpec,omitempty"`
	ReadinessInitialDelaySeconds *int32         `json:"readinessDelaySec,omitempty"`
	LivenessInitialDelaySeconds  *int32         `json:"livenessDelaySec,omitempty"`
	MultiAZ
}

type VolumeSpec struct {
	// EmptyDir represents a temporary directory that shares a pod's lifetime.
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath represents a pre-existing file or directory on the host machine
	// that is directly exposed to the container.
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim represents a reference to a PersistentVolumeClaim.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type ResourceSpecRequirements struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ResourcesSpec struct {
	Limits   *ResourceSpecRequirements `json:"limits,omitempty"`
	Requests *ResourceSpecRequirements `json:"requests,omitempty"`
}

type SecretsSpec struct {
	Users       string `json:"users,omitempty"`
	SSL         string `json:"ssl,omitempty"`
	SSLInternal string `json:"sslInternal,omitempty"`
}

type MongosSpec struct {
	*ResourcesSpec `json:"resources,omitempty"`
	Port           int32 `json:"port,omitempty"`
	HostPort       int32 `json:"hostPort,omitempty"`
}

type MongodSpec struct {
	Net                *MongodSpecNet                `json:"net,omitempty"`
	AuditLog           *MongodSpecAuditLog           `json:"auditLog,omitempty"`
	OperationProfiling *MongodSpecOperationProfiling `json:"operationProfiling,omitempty"`
	Replication        *MongodSpecReplication        `json:"replication,omitempty"`
	Security           *MongodSpecSecurity           `json:"security,omitempty"`
	SetParameter       *MongodSpecSetParameter       `json:"setParameter,omitempty"`
	Storage            *MongodSpecStorage            `json:"storage,omitempty"`
}

type MongodSpecNet struct {
	Port     int32 `json:"port,omitempty"`
	HostPort int32 `json:"hostPort,omitempty"`
}

type MongodSpecReplication struct {
	OplogSizeMB int `json:"oplogSizeMB,omitempty"`
}

// MongodChiperMode is a cipher mode used by Data-at-Rest Encryption
type MongodChiperMode string

const (
	MongodChiperModeUnset MongodChiperMode = ""
	MongodChiperModeCBC                    = "AES256-CBC"
	MongodChiperModeGCM                    = "AES256-GCM"
)

type MongodSpecSecurity struct {
	RedactClientLogData  bool             `json:"redactClientLogData,omitempty"`
	EnableEncryption     *bool            `json:"enableEncryption,omitempty"`
	EncryptionKeySecret  string           `json:"encryptionKeySecret,omitempty"`
	EncryptionCipherMode MongodChiperMode `json:"encryptionCipherMode,omitempty"`
}

type MongodSpecSetParameter struct {
	TTLMonitorSleepSecs                   int `json:"ttlMonitorSleepSecs,omitempty"`
	WiredTigerConcurrentReadTransactions  int `json:"wiredTigerConcurrentReadTransactions,omitempty"`
	WiredTigerConcurrentWriteTransactions int `json:"wiredTigerConcurrentWriteTransactions,omitempty"`
}

type StorageEngine string

var (
	StorageEngineWiredTiger StorageEngine = "wiredTiger"
	StorageEngineInMemory   StorageEngine = "inMemory"
	StorageEngineMMAPv1     StorageEngine = "mmapv1"
)

type MongodSpecStorage struct {
	Engine         StorageEngine         `json:"engine,omitempty"`
	DirectoryPerDB bool                  `json:"directoryPerDB,omitempty"`
	SyncPeriodSecs int                   `json:"syncPeriodSecs,omitempty"`
	InMemory       *MongodSpecInMemory   `json:"inMemory,omitempty"`
	MMAPv1         *MongodSpecMMAPv1     `json:"mmapv1,omitempty"`
	WiredTiger     *MongodSpecWiredTiger `json:"wiredTiger,omitempty"`
}

type MongodSpecMMAPv1 struct {
	NsSize     int  `json:"nsSize,omitempty"`
	Smallfiles bool `json:"smallfiles,omitempty"`
}

type WiredTigerCompressor string

var (
	WiredTigerCompressorNone   WiredTigerCompressor = "none"
	WiredTigerCompressorSnappy WiredTigerCompressor = "snappy"
	WiredTigerCompressorZlib   WiredTigerCompressor = "zlib"
)

type MongodSpecWiredTigerEngineConfig struct {
	CacheSizeRatio      float64               `json:"cacheSizeRatio,omitempty"`
	DirectoryForIndexes bool                  `json:"directoryForIndexes,omitempty"`
	JournalCompressor   *WiredTigerCompressor `json:"journalCompressor,omitempty"`
}

type MongodSpecWiredTigerCollectionConfig struct {
	BlockCompressor *WiredTigerCompressor `json:"blockCompressor,omitempty"`
}

type MongodSpecWiredTigerIndexConfig struct {
	PrefixCompression bool `json:"prefixCompression,omitempty"`
}

type MongodSpecWiredTiger struct {
	CollectionConfig *MongodSpecWiredTigerCollectionConfig `json:"collectionConfig,omitempty"`
	EngineConfig     *MongodSpecWiredTigerEngineConfig     `json:"engineConfig,omitempty"`
	IndexConfig      *MongodSpecWiredTigerIndexConfig      `json:"indexConfig,omitempty"`
}

type MongodSpecInMemoryEngineConfig struct {
	InMemorySizeRatio float64 `json:"inMemorySizeRatio,omitempty"`
}

type MongodSpecInMemory struct {
	EngineConfig *MongodSpecInMemoryEngineConfig `json:"engineConfig,omitempty"`
}

type AuditLogDestination string

var AuditLogDestinationFile AuditLogDestination = "file"

type AuditLogFormat string

var (
	AuditLogFormatBSON AuditLogFormat = "BSON"
	AuditLogFormatJSON AuditLogFormat = "JSON"
)

type MongodSpecAuditLog struct {
	Destination AuditLogDestination `json:"destination,omitempty"`
	Format      AuditLogFormat      `json:"format,omitempty"`
	Filter      string              `json:"filter,omitempty"`
}

type OperationProfilingMode string

const (
	OperationProfilingModeAll    OperationProfilingMode = "all"
	OperationProfilingModeSlowOp OperationProfilingMode = "slowOp"
)

type MongodSpecOperationProfiling struct {
	Mode              OperationProfilingMode `json:"mode,omitempty"`
	SlowOpThresholdMs int                    `json:"slowOpThresholdMs,omitempty"`
	RateLimit         int                    `json:"rateLimit,omitempty"`
}

type BackupCoordinatorSpec struct {
	Resources                   *corev1.ResourceRequirements `json:"resources,omitempty"`
	StorageClass                string                       `json:"storageClass,omitempty"`
	EnableClientsLogging        bool                         `json:"enableClientsLogging,omitempty"`
	LivenessInitialDelaySeconds *int32                       `json:"livenessDelaySec,omitempty"`
	MultiAZ
}

type BackupDestinationType string

var (
	BackupDestinationS3   BackupDestinationType = "s3"
	BackupDestinationFile BackupDestinationType = "file"
)

type BackupCompressionType string

var (
	BackupCompressionGzip BackupCompressionType = "gzip"
)

type BackupTaskSpec struct {
	Name            string                `json:"name"`
	Enabled         bool                  `json:"enabled"`
	Schedule        string                `json:"schedule,omitempty"`
	StorageName     string                `json:"storageName,omitempty"`
	CompressionType BackupCompressionType `json:"compressionType,omitempty"`
}

type BackupSpec struct {
	Enabled          bool                               `json:"enabled"`
	Debug            bool                               `json:"debug"`
	RestartOnFailure *bool                              `json:"restartOnFailure,omitempty"`
	Coordinator      BackupCoordinatorSpec              `json:"coordinator,omitempty"`
	Storages         map[string]dbaas.BackupStorageSpec `json:"storages,omitempty"`
	Image            string                             `json:"image,omitempty"`
	Tasks            []BackupTaskSpec                   `json:"tasks,omitempty"`
}

type Arbiter struct {
	Enabled bool  `json:"enabled"`
	Size    int32 `json:"size"`
	MultiAZ
}

type Expose struct {
	Enabled    bool               `json:"enabled"`
	ExposeType corev1.ServiceType `json:"exposeType,omitempty"`
}

type Platform string

const (
	PlatformUndef      Platform = ""
	PlatformKubernetes          = "kubernetes"
	PlatformOpenshift           = "openshift"
)

const AffinityTopologyKeyOff = "none"

var affinityValidTopologyKeys = map[string]struct{}{
	AffinityTopologyKeyOff:                     struct{}{},
	"kubernetes.io/hostname":                   struct{}{},
	"failure-domain.beta.kubernetes.io/zone":   struct{}{},
	"failure-domain.beta.kubernetes.io/region": struct{}{},
}

var defaultAffinityTopologyKey = "kubernetes.io/hostname"

// ServerVersion represents info about k8s / openshift server version
type ServerVersion struct {
	Platform Platform
	Info     k8sversion.Info
}

func (cr *PerconaServerMongoDB) UpdateWith(rsName string, f *pflag.FlagSet, s3 *dbaas.BackupStorageSpec) (err error) {
	if _, ok := cr.Spec.Backup.Storages[dbaas.DefaultBcpStorageName]; !ok && s3 != nil {
		if cr.Spec.Backup.Storages == nil {
			cr.Spec.Backup.Storages = make(map[string]dbaas.BackupStorageSpec)
		}

		cr.Spec.Backup.Storages[dbaas.DefaultBcpStorageName] = *s3
	}

	size, err := f.GetInt32("replset-size")
	if err != nil {
		return errors.New("undefined `replset-size`")
	}

	if size == 0 {
		return nil
	}

	for _, rs := range cr.Spec.Replsets {
		if rs.Name == rsName {
			rs.Size = size

			return nil
		}
	}

	return errors.Errorf("unknown replica set '%s'", rsName)
}

func NewReplSet(name string, f *pflag.FlagSet) (*ReplsetSpec, error) {
	rs := &ReplsetSpec{
		Name: name,
	}

	volSizeFlag, err := f.GetString("storage-size")
	if err != nil {
		return nil, errors.New("undefined `storage-size`")
	}
	volSize, err := resource.ParseQuantity(volSizeFlag)
	if err != nil {
		return nil, errors.Wrap(err, "storage-size")
	}

	rs.VolumeSpec = &VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volSize},
			},
		},
	}

	volClassNameFlag, err := f.GetString("storage-class")
	if err != nil {
		return nil, errors.New("undefined `storage-class`")
	}

	if volClassNameFlag != "" {
		rs.VolumeSpec.PersistentVolumeClaim.StorageClassName = &volClassNameFlag
	}

	rs.Size, err = f.GetInt32("replset-size")
	if err != nil {
		return nil, errors.New("undefined `replset-size`")
	}

	psmdbCPU, err := f.GetString("request-cpu")
	if err != nil {
		return nil, errors.New("undefined `request-cpu`")
	}
	_, err = resource.ParseQuantity(psmdbCPU)
	if err != nil {
		return nil, errors.Wrap(err, "request-cpu")
	}
	psmdbMemory, err := f.GetString("request-mem")
	if err != nil {
		return nil, errors.New("undefined `request-mem`")
	}
	_, err = resource.ParseQuantity(psmdbMemory)
	if err != nil {
		return nil, errors.Wrap(err, "request-mem")
	}
	rs.Resources = &ResourcesSpec{
		Requests: &ResourceSpecRequirements{
			CPU:    psmdbCPU,
			Memory: psmdbMemory,
		},
	}

	psmdbtpk, err := f.GetString("anti-affinity-key")
	if err != nil {
		return nil, errors.New("undefined `anti-affinity-key`")
	}
	if _, ok := affinityValidTopologyKeys[psmdbtpk]; !ok {
		return nil, errors.Errorf("invalid `anti-affinity-key` value: %s", psmdbtpk)
	}
	rs.Affinity = &PodAffinity{
		TopologyKey: &psmdbtpk,
	}

	return rs, nil
}


// Upgrade upgrades culster with given images
func (cr *PerconaServerMongoDB) Upgrade(imgs map[string]string) {
	if img, ok := imgs["psmdb"]; ok {
		cr.Spec.Image = img
	}
	if img, ok := imgs["backup"]; ok {
		cr.Spec.Backup.Image = img
	}
}

func (cr *PerconaServerMongoDB) SetNew(clusterName, rsName string, f *pflag.FlagSet, s3 *dbaas.BackupStorageSpec) (err error) {
	cr.ObjectMeta.Name = clusterName
	cr.setDefaults()

	rs, err := NewReplSet(rsName, f)
	if err != nil {
		return errors.Wrap(err, "new replset")
	}

	cr.Spec.Replsets = []*ReplsetSpec{
		rs,
	}

	cr.Spec.Backup, err = cr.createBackup(f)
	if err != nil {
		return errors.Wrap(err, "backup spec")
	}

	if s3 != nil {
		cr.Spec.Backup.Storages = map[string]dbaas.BackupStorageSpec{
			dbaas.DefaultBcpStorageName: *s3,
		}
	}

	switch p {
	case dbaas.PlatformMinishift, dbaas.PlatformMinikube:
		none := AffinityTopologyKeyOff
		for i, _ := range cr.Spec.Replsets {
			cr.Spec.Replsets[i].Resources = nil
			cr.Spec.Replsets[i].MultiAZ.Affinity.TopologyKey = &none
		}
	}

	return nil
}

func (cr *PerconaServerMongoDB) setDefaults() {
	cr.TypeMeta.APIVersion = "psmdb.percona.com/v1"
	cr.TypeMeta.Kind = "PerconaServerMongoDB"

	cr.Spec.Secrets = &SecretsSpec{
		Users: cr.Name + "-secrets",
	}

	cr.Spec.Image = "percona/percona-server-mongodb-operator:1.0.0-mongod4.0.9"
}

func (cr *PerconaServerMongoDB) createBackup(f *pflag.FlagSet) (BackupSpec, error) {
	t := true
	volSize, err := resource.ParseQuantity("1Gi")
	if err != nil {
		return BackupSpec{}, errors.Wrap(err, "coordinator Storage")
	}
	cpu, err := resource.ParseQuantity("100m")
	if err != nil {
		return BackupSpec{}, errors.Wrap(err, "coordinator CPU")
	}
	mem, err := resource.ParseQuantity("0.1G")
	if err != nil {
		return BackupSpec{}, errors.Wrap(err, "coordinator Memory")
	}
	memlim, err := resource.ParseQuantity("0.2G")
	if err != nil {
		return BackupSpec{}, errors.Wrap(err, "coordinator Memory limit")
	}
	bcp := BackupSpec{
		Enabled:          true,
		RestartOnFailure: &t,
		Debug:            true,
		Image:            "perconalab/percona-server-mongodb-operator:1.1.0-backup",
		Coordinator: BackupCoordinatorSpec{
			EnableClientsLogging: true,
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    cpu,
					corev1.ResourceMemory: memlim,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:     cpu,
					corev1.ResourceMemory:  mem,
					corev1.ResourceStorage: volSize,
				},
			},
		},
	}

	return bcp, nil
}

// PerconaServerMongoDBBackupSpec defines the desired state of PerconaServerMongoDBBackup
type PerconaServerMongoDBBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	PSMDBCluster string `json:"psmdbCluster,omitempty"`
	StorageName  string `json:"storageName,omitempty"`
}

type PerconaSMDBStatusState string

const (
	StateRequested PerconaSMDBStatusState = "requested"
	StateRejected                         = "rejected"
	StateReady                            = "ready"
)

// PerconaServerMongoDBBackupStatus defines the observed state of PerconaServerMongoDBBackup
type PerconaServerMongoDBBackupStatus struct {
	State         PerconaSMDBStatusState     `json:"state,omitempty"`
	StartAt       *metav1.Time               `json:"start,omitempty"`
	CompletedAt   *metav1.Time               `json:"completed,omitempty"`
	LastScheduled *metav1.Time               `json:"lastscheduled,omitempty"`
	Destination   string                     `json:"destination,omitempty"`
	StorageName   string                     `json:"storageName,omitempty"`
	S3            *dbaas.BackupStorageS3Spec `json:"s3,omitempty"`
}

// PerconaServerMongoDBBackup is the Schema for the perconaservermongodbbackups API
type PerconaServerMongoDBBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerconaServerMongoDBBackupSpec   `json:"spec,omitempty"`
	Status PerconaServerMongoDBBackupStatus `json:"status,omitempty"`
}

// PerconaServerMongoDBBackupList contains a list of PerconaServerMongoDBBackup
type PerconaServerMongoDBBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerconaServerMongoDBBackup `json:"items"`
}

func (b *PerconaServerMongoDBBackup) SetNew(name, cluster, storage string) {
	b.TypeMeta.APIVersion = "psmdb.percona.com/v1"
	b.TypeMeta.Kind = "PerconaServerMongoDBBackup"

	b.ObjectMeta.Name = name
	b.Spec.PSMDBCluster = cluster
	b.Spec.StorageName = storage
}

// PerconaServerMongoDBRestoreSpec defines the desired state of PerconaServerMongoDBRestore
type PerconaServerMongoDBRestoreSpec struct {
	BackupName  string `json:"backupName,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	Destination string `json:"destination,omitempty"`
	StorageName string `json:"storageName,omitempty"`
}

// PerconaSMDBRestoreStatusState is for restore status states
type PerconaSMDBRestoreStatusState string

const (
	RestoreStateRequested PerconaSMDBRestoreStatusState = "requested"
	RestoreStateReady     PerconaSMDBRestoreStatusState = "ready"
	RestoreStateRejected  PerconaSMDBRestoreStatusState = "rejected"
)

// PerconaServerMongoDBRestoreStatus defines the observed state of PerconaServerMongoDBRestore
type PerconaServerMongoDBRestoreStatus struct {
	State PerconaSMDBRestoreStatusState `json:"state,omitempty"`
}

// PerconaServerMongoDBRestore is the Schema for the perconaservermongodbrestores API
type PerconaServerMongoDBRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerconaServerMongoDBRestoreSpec   `json:"spec,omitempty"`
	Status PerconaServerMongoDBRestoreStatus `json:"status,omitempty"`
}

// PerconaServerMongoDBRestoreList contains a list of PerconaServerMongoDBRestore
type PerconaServerMongoDBRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerconaServerMongoDBRestore `json:"items"`
}

func (b *PerconaServerMongoDBRestore) SetNew(name, cluster, backup string) {
	b.TypeMeta.APIVersion = "psmdb.percona.com/v1"
	b.TypeMeta.Kind = "PerconaServerMongoDBRestore"

	b.ObjectMeta.Name = name
	b.Spec.ClusterName = cluster
	b.Spec.BackupName = backup
}
