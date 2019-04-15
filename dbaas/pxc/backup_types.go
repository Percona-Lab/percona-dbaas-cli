package pxc

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PerconaXtraDBBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PerconaXtraDBBackup `json:"items"`
}

type PerconaXtraDBBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PXCBackupSpec   `json:"spec"`
	Status            PXCBackupStatus `json:"status,omitempty"`
}

type PXCBackupSpec struct {
	PXCCluster  string `json:"pxcCluster"`
	StorageName string `json:"storageName,omitempty"`
}

type PXCBackupStatus struct {
	State         PXCBackupState       `json:"state,omitempty"`
	CompletedAt   *metav1.Time         `json:"completed,omitempty"`
	LastScheduled *metav1.Time         `json:"lastscheduled,omitempty"`
	Destination   string               `json:"destination,omitempty"`
	StorageName   string               `json:"storageName,omitempty"`
	S3            *BackupStorageS3Spec `json:"s3,omitempty"`
}

type PXCBackupState string

const (
	BackupStarting  PXCBackupState = "Starting"
	BackupRunning                  = "Running"
	BackupFailed                   = "Failed"
	BackupSucceeded                = "Succeeded"
)

func (b *PerconaXtraDBBackup) SetNew(name, cluster, storage string) {
	b.TypeMeta.APIVersion = "pxc.percona.com/v1alpha1"
	b.TypeMeta.Kind = "PerconaXtraDBBackup"

	b.ObjectMeta.Name = name
	b.Spec.PXCCluster = cluster
	b.Spec.StorageName = storage
}
