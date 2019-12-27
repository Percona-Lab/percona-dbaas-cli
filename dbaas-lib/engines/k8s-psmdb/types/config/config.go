package config

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
)

type ClusterConfig struct {
	OperatorImage    string            `json:"operatorImage,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	Name             string
	Pause            bool                          `json:"pause,omitempty"`
	Platform         Platform                      `json:"platform,omitempty"`
	Image            string                        `json:"image,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	RunUID           int64                         `json:"runUid,omitempty"`
	UnsafeConf       bool                          `json:"allowUnsafeConfigurations"`
	Mongod           MongodSpec                    `json:"mongod,omitempty"`
	Replsets         []ReplsetSpec                 `json:"replsets,omitempty"`
	Secrets          SecretsSpec                   `json:"secrets,omitempty"`
	Backup           BackupSpec                    `json:"backup,omitempty"`
	ImagePullPolicy  corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	PMM              PMMSpec                       `json:"pmm,omitempty"`
	BrokerInstance   string                        `json:"brokerInstance,omitempty"`
}

type ReplsetSpec struct {
	Resources                    ResourcesSpec `json:"resources,omitempty"`
	Name                         string        `json:"name"`
	Size                         int32         `json:"size"`
	ClusterRole                  ClusterRole   `json:"clusterRole,omitempty"`
	Arbiter                      Arbiter       `json:"arbiter,omitempty"`
	Expose                       Expose        `json:"expose,omitempty"`
	VolumeSpec                   VolumeSpec    `json:"volumeSpec,omitempty"`
	ReadinessInitialDelaySeconds int32         `json:"readinessDelaySec,omitempty"`
	LivenessInitialDelaySeconds  int32         `json:"livenessDelaySec,omitempty"`
	MultiAZ
}

type Platform string

type ClusterRole string

type VolumeSpec struct {
	// EmptyDir represents a temporary directory that shares a pod's lifetime.
	EmptyDir corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath represents a pre-existing file or directory on the host machine
	// that is directly exposed to the container.
	HostPath corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim represents a reference to a PersistentVolumeClaim.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	PersistentVolumeClaim corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type ResourceSpecRequirements struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ResourcesSpec struct {
	Limits   ResourceSpecRequirements `json:"limits,omitempty"`
	Requests ResourceSpecRequirements `json:"requests,omitempty"`
}

type PMMSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerHost string `json:"serverHost,omitempty"`
	Image      string `json:"image,omitempty"`
}

type MultiAZ struct {
	Affinity            PodAffinity             `json:"affinity,omitempty"`
	NodeSelector        map[string]string       `json:"nodeSelector,omitempty"`
	Tolerations         []corev1.Toleration     `json:"tolerations,omitempty"`
	PriorityClassName   string                  `json:"priorityClassName,omitempty"`
	Annotations         map[string]string       `json:"annotations,omitempty"`
	Labels              map[string]string       `json:"labels,omitempty"`
	PodDisruptionBudget PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type PodAffinity struct {
	TopologyKey string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    corev1.Affinity `json:"advanced,omitempty"`
}

type MongodSpec struct {
	Net                MongodSpecNet                `json:"net,omitempty"`
	AuditLog           MongodSpecAuditLog           `json:"auditLog,omitempty"`
	OperationProfiling MongodSpecOperationProfiling `json:"operationProfiling,omitempty"`
	Replication        MongodSpecReplication        `json:"replication,omitempty"`
	Security           MongodSpecSecurity           `json:"security,omitempty"`
	SetParameter       MongodSpecSetParameter       `json:"setParameter,omitempty"`
	Storage            MongodSpecStorage            `json:"storage,omitempty"`
}

type MongodSpecNet struct {
	Port     int32 `json:"port,omitempty"`
	HostPort int32 `json:"hostPort,omitempty"`
}

type MongodSpecReplication struct {
	OplogSizeMB int `json:"oplogSizeMB,omitempty"`
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

type SecretsSpec struct {
	Users       string `json:"users,omitempty"`
	SSL         string `json:"ssl,omitempty"`
	SSLInternal string `json:"sslInternal,omitempty"`
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
	Enabled          bool                             `json:"enabled"`
	Debug            bool                             `json:"debug"`
	RestartOnFailure *bool                            `json:"restartOnFailure,omitempty"`
	Coordinator      BackupCoordinatorSpec            `json:"coordinator,omitempty"`
	Storages         map[string]k8s.BackupStorageSpec `json:"storages,omitempty"`
	Image            string                           `json:"image,omitempty"`
	Tasks            []BackupTaskSpec                 `json:"tasks,omitempty"`
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
