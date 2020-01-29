package v130

import "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

// Bundle is bundle object for version 1.2.0
var Bundle = []k8s.BundleObject{
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaxtradbclusters.pxc.percona.com",
		Data: `
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
  versions:
    - name: v1
      storage: true
      served: true
    - name: v1-2-0
      storage: false
      served: true
    - name: v1-3-0
      storage: false
      served: true
    - name: v1alpha1
      storage: false
      served: true
  additionalPrinterColumns:
    - name: Endpoint
      type: string
      JSONPath: .status.host
    - name: Status
      type: string
      JSONPath: .status.state
    - name: PXC
      type: string
      description: Ready pxc nodes
      JSONPath: .status.pxc.ready
    - name: proxysql
      type: string
      description: Ready pxc nodes
      JSONPath: .status.proxysql.ready
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
  subresources:
    status: {}
`,
	},
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaxtradbclusterbackups.pxc.percona.com",
		Data: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: perconaxtradbclusterbackups.pxc.percona.com
spec:
  group: pxc.percona.com
  names:
    kind: PerconaXtraDBClusterBackup
    listKind: PerconaXtraDBClusterBackupList
    plural: perconaxtradbclusterbackups
    singular: perconaxtradbclusterbackup
    shortNames:
    - pxc-backup
    - pxc-backups
  scope: Namespaced
  versions:
    - name: v1
      storage: true
      served: true
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
  subresources:
    status: {}
`,
	},
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaxtradbclusterrestores.pxc.percona.com",
		Data: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: perconaxtradbclusterrestores.pxc.percona.com
spec:
  group: pxc.percona.com
  names:
    kind: PerconaXtraDBClusterRestore
    listKind: PerconaXtraDBClusterRestoreList
    plural: perconaxtradbclusterrestores
    singular: perconaxtradbclusterrestore
    shortNames:
    - pxc-restore
    - pxc-restores
  scope: Namespaced
  versions:
    - name: v1
      storage: true
      served: true
  additionalPrinterColumns:
    - name: Cluster
      type: string
      description: Cluster name
      JSONPath: .spec.pxcCluster
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
  subresources:
    status: {}
`,
	},
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaxtradbbackups.pxc.percona.com",
		Data: `
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
    shortNames: []
  scope: Namespaced
  versions:
    - name: v1alpha1
      storage: true
      served: true
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
`,
	},
	{
		Kind: "Role",
		Name: "percona-xtradb-cluster-operator",
		Data: `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: percona-xtradb-cluster-operator
rules:
- apiGroups:
  - pxc.percona.com
  resources:
  - perconaxtradbclusters
  - perconaxtradbclusters/status
  - perconaxtradbclusterbackups
  - perconaxtradbclusterbackups/status
  - perconaxtradbclusterrestores
  - perconaxtradbclusterrestores/status
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
- apiGroups:
  - certmanager.k8s.io
  resources:
  - issuers
  - certificates
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
  - deletecollection
`,
	},
	{
		Kind: "RoleBinding",
		Name: "default-account-percona-xtradb-cluster-operator",
		Data: `
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
`,
	},
	{
		Kind: "Deployment",
		Name: "percona-xtradb-cluster-operator",
		Data: `
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
          image: {{image}}
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
	},
}
