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
	b.TypeMeta.APIVersion = "pxc.percona.com/v1"
	b.TypeMeta.Kind = "PerconaXtraDBClusterBackup"

	b.ObjectMeta.Name = name
	b.Spec.PXCCluster = cluster
	b.Spec.StorageName = storage
}
