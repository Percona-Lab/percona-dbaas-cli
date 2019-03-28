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
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

var objects = map[Version]dbaas.Objects{
	Version030: dbaas.Objects{
		CR: `
apiVersion: "pxc.percona.com/v1alpha1"
kind: "PerconaXtraDBCluster"
metadata:
	name: "{{.ClusterName}}"
	finalizers:
	- delete-pxc-pods-in-order
#    - delete-proxysql-pvc
#    - delete-pxc-pvc
spec:
	secretsName: my-cluster-secrets
	pxc:
	size: {{.PXC.Size}}
	allowUnsafeConfigurations: false
	image: percona/percona-xtradb-cluster-operator:0.3.0-pxc
#    configuration: |
#      [mysqld]
#      wsrep_debug=ON
#      [sst]
#      wsrep_debug=ON
#    imagePullSecrets:
#      - name: private-registry-credentials
#    priorityClassName: high-priority
#    annotations:
#      iam.amazonaws.com/role: role-arn
#    labels:
#      rack: rack-22
	resources:
		requests:
		memory: {{.PXC.Requests.Memory}}
		cpu: {{.PXC.Requests.CPU}}
#      limits:
#        memory: 1G
#        cpu: "1"
#    nodeSelector:
#      disktype: ssd
	affinity:
		antiAffinityTopologyKey: "kubernetes.io/hostname"
#      advanced:
#        nodeAffinity:
#          requiredDuringSchedulingIgnoredDuringExecution:
#            nodeSelectorTerms:
#            - matchExpressions:
#              - key: kubernetes.io/e2e-az-name
#                operator: In
#                values:
#                - e2e-az1
#                - e2e-az2
#    tolerations: 
#    - key: "node.alpha.kubernetes.io/unreachable"
#      operator: "Exists"
#      effect: "NoExecute"
#      tolerationSeconds: 6000
	podDisruptionBudget:
		maxUnavailable: 1
#      minAvailable: 0
	volumeSpec:
#      emptyDir: {}
#      hostPath:
#        path: /data
#        type: Directory
		persistentVolumeClaim:
#        storageClassName: standard
#        accessModes: [ "ReadWriteOnce" ]
		resources:
			requests:
			storage: {{.PXC.Storage}}
	proxysql:
	enabled: true
	size: {{.Proxy.Size}}
	image: percona/percona-xtradb-cluster-operator:0.3.0-proxysql
#    imagePullSecrets:
#      - name: private-registry-credentials
#    annotations:
#      iam.amazonaws.com/role: role-arn
#    labels:
#      rack: rack-22
	resources:
		requests:
		memory: {{.Proxy.Requests.Memory}}
		cpu: {{.Proxy.Requests.CPU}}
#      limits:
#        memory: 1G
#        cpu: 700m
#    priorityClassName: high-priority
#    nodeSelector:
#      disktype: ssd
	affinity:
		antiAffinityTopologyKey: "kubernetes.io/hostname"
#      advanced:
#        nodeAffinity:
#          requiredDuringSchedulingIgnoredDuringExecution:
#            nodeSelectorTerms:
#            - matchExpressions:
#              - key: kubernetes.io/e2e-az-name
#                operator: In
#                values:
#                - e2e-az1
#                - e2e-az2
#    tolerations:
#    - key: "node.alpha.kubernetes.io/unreachable"
#      operator: "Exists"
#      effect: "NoExecute"
#      tolerationSeconds: 6000
	volumeSpec:
#      emptyDir: {}
#      hostPath:
#        path: /data
#        type: Directory
		persistentVolumeClaim:
#        storageClassName: standard
#        accessModes: [ "ReadWriteOnce" ]
		resources:
			requests:
			storage: {{.Proxy.Storage}}
	podDisruptionBudget:
		maxUnavailable: 1
#      minAvailable: 0
	pmm:
	enabled: false
	image: perconalab/pmm-client:1.17.1
	serverHost: monitoring-service
	serverUser: pmm
#   backup:
#     image: percona/percona-xtradb-cluster-operator:0.3.0-backup
# #    imagePullSecrets:
# #      - name: private-registry-credentials
#     storages:
#       s3-us-west:
#         type: s3
#         s3:
#           bucket: S3-BACKUP-BUCKET-NAME-HERE
#           credentialsSecret: my-cluster-name-backup-s3
#           region: us-west-2
#       fs-pvc:
#         type: filesystem
#         volume:
#           persistentVolumeClaim:
# #            storageClassName: standard
#             accessModes: [ "ReadWriteOnce" ]
#             resources:
#               requests:
#                 storage: 6Gi
#     schedule:
#       - name: "sat-night-backup"
#         schedule: "0 0 * * 6"
#         keep: 3
#         storageName: s3-us-west
#       - name: "daily-backup"
#         schedule: "0 0 * * *"
#         keep: 5
#         storageName: fs-pvc		
		`,

		Bundle: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
	name: perconaxtradbclusters.pxc.percona.com
spec:
	group: pxc.percona.com
	names:
	kind: PerconaXtraDBCluster
	listKind: PerconaXtraDBClusterList
	plural: perconaxtradbclusters
	singular: perconaxtradbcluster
	shortNames:
	- pxc
	- pxcs
	scope: Namespaced
	version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
	name: perconaxtradbbackups.pxc.percona.com
spec:
	group: pxc.percona.com
	names:
	kind: PerconaXtraDBBackup
	listKind: PerconaXtraDBBackupList
	plural: perconaxtradbbackups
	singular: perconaxtradbbackup
	shortNames:
	- pxc-backup
	- pxc-backups
	scope: Namespaced
	version: v1alpha1
	additionalPrinterColumns:
	- name: Cluster
		type: string
		description: Cluster name
		JSONPath: .spec.pxcCluster
	- name: Storage
		type: string
		description: Storage name from pxc spec
		JSONPath: .status.storageName
	- name: Destination
		type: string
		description: Backup destination
		JSONPath: .status.destination
	- name: Status
		type: string
		description: Job status
		JSONPath: .status.state
	- name: Completed
		description: Completed time
		type: date
		JSONPath: .status.completed
	- name: Age
		type: date
		JSONPath: .metadata.creationTimestamp
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
	name: percona-xtradb-cluster-operator
rules:
- apiGroups:
	- pxc.percona.com
	resources:
	- perconaxtradbclusters
	- perconaxtradbbackups
	verbs:
	- get
	- list
	- watch
	- create
	- update
	- patch
	- delete
- apiGroups:
	- ""
	resources:
	- pods
	- pods/exec
	- configmaps
	- services
	- persistentvolumeclaims
	- secrets
	verbs:
	- get
	- list
	- watch
	- create
	- update
	- patch
	- delete
- apiGroups:
	- apps
	resources:
	- deployments
	- replicasets
	- statefulsets
	verbs:
	- get
	- list
	- watch
	- create
	- update
	- patch
	- delete
- apiGroups:
	- batch
	resources:
	- jobs
	- cronjobs
	verbs:
	- get
	- list
	- watch
	- create
	- update
	- patch
	- delete
- apiGroups:
	- policy
	resources:
	- poddisruptionbudgets
	verbs:
	- get
	- list
	- watch
	- create
	- update
	- patch
	- delete
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
	name: default-account-percona-xtradb-cluster-operator
subjects:
- kind: ServiceAccount
	name: default
roleRef:
	kind: Role
	name: percona-xtradb-cluster-operator
	apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
	name: percona-xtradb-cluster-operator
spec:
	replicas: 1
	selector:
	matchLabels:
		name: percona-xtradb-cluster-operator
	template:
	metadata:
		labels:
		name: percona-xtradb-cluster-operator
	spec:
		containers:
		- name: percona-xtradb-cluster-operator
			image: percona/percona-xtradb-cluster-operator:0.3.0
			ports:
			- containerPort: 60000
			name: metrics
			command:
			- percona-xtradb-cluster-operator
			imagePullPolicy: Always
			env:
			- name: WATCH_NAMESPACE
				valueFrom:
				fieldRef:
					fieldPath: metadata.namespace
			- name: OPERATOR_NAME
				value: "percona-xtradb-cluster-operator"	
	`,
		Secrets: `
apiVersion: v1
kind: Secret
metadata:
	name: my-cluster-secrets
type: Opaque
data:
	root: cm9vdF9wYXNzd29yZA==
	xtrabackup: YmFja3VwX3Bhc3N3b3Jk
	monitor: bW9uaXRvcg==
	clustercheck: Y2x1c3RlcmNoZWNrcGFzc3dvcmQ=
	proxyadmin: YWRtaW5fcGFzc3dvcmQ=
	pmmserver: c3VwYXxefHBheno=	
	`,
	},
}
