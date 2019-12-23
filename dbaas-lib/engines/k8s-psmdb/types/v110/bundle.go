package v110

import "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

var Bundle = []k8s.BundleObject{
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaservermongodbs.psmdb.percona.com",
		Data: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: perconaservermongodbs.psmdb.percona.com
spec:
  group: psmdb.percona.com
  names:
    kind: PerconaServerMongoDB
    listKind: PerconaServerMongoDBList
    plural: perconaservermongodbs
    singular: perconaservermongodb
    shortNames:
    - psmdb
  scope: Namespaced
  versions:
    - name: v1
      storage: true
      served: true
    - name: v1alpha1
      storage: false
      served: true
  additionalPrinterColumns:
    - name: Status
      type: string
      JSONPath: .status.state
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
  subresources:
    status: {}
`,
	},
	{
		Kind: "CustomResourceDefinition",
		Name: "perconaservermongodbbackups.psmdb.percona.com",
		Data: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: perconaservermongodbbackups.psmdb.percona.com
spec:
  group: psmdb.percona.com
  names:
    kind: PerconaServerMongoDBBackup
    listKind: PerconaServerMongoDBBackupList
    plural: perconaservermongodbbackups
    singular: perconaservermongodbbackup
    shortNames:
    - psmdb-backup
  scope: Namespaced
  versions:
    - name: v1
      storage: true
      served: true
  additionalPrinterColumns:
    - name: Cluster
      type: string
      description: Cluster name
      JSONPath: .spec.psmdbCluster
    - name: Storage
      type: string
      description: Storage name from psmdb spec
      JSONPath: .spec.storageName
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
		Name: "perconaservermongodbrestores.psmdb.percona.com",
		Data: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: perconaservermongodbrestores.psmdb.percona.com
spec:
  group: psmdb.percona.com
  names:
    kind: PerconaServerMongoDBRestore
    listKind: PerconaServerMongoDBRestoreList
    plural: perconaservermongodbrestores
    singular: perconaservermongodbrestore
    shortNames:
    - psmdb-restore
  scope: Namespaced
  versions:
    - name: v1
      storage: true
      served: true
  additionalPrinterColumns:
    - name: Cluster
      type: string
      description: Cluster name
      JSONPath: .spec.clusterName
    - name: Status
      type: string
      description: Job status
      JSONPath: .status.state
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
  subresources:
    status: {}
`,
	},
	{
		Kind: "Role",
		Name: "percona-server-mongodb-operator",
		Data: `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: percona-server-mongodb-operator
rules:
  - apiGroups:
    - psmdb.percona.com
    resources:
    - perconaservermongodbs
    - perconaservermongodbs/status
    - perconaservermongodbbackups
    - perconaservermongodbbackups/status
    - perconaservermongodbrestores
    - perconaservermongodbrestores/status
    verbs:
    - get
    - list
    - update
    - watch
    - create
  - apiGroups:
    - ""
    resources:
    - pods
    - pods/exec
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
		Name: "default-account-percona-server-mongodb-operator",
		Data: `
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: default-account-percona-server-mongodb-operator
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: percona-server-mongodb-operator
  apiGroup: rbac.authorization.k8s.io
`,
	},
	{
		Kind: "Deployment",
		Name: "percona-server-mongodb-operator",
		Data: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: percona-server-mongodb-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: percona-server-mongodb-operator
  template:
    metadata:
      labels:
        name: percona-server-mongodb-operator
    spec:
      containers:
        - name: percona-server-mongodb-operator
          image: {{image}}
          ports:
          - containerPort: 60000
            name: metrics
          command:
          - percona-server-mongodb-operator
          imagePullPolicy: Always
          env:
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: OPERATOR_NAME
              value: percona-server-mongodb-operator
            - name: RESYNC_PERIOD
              value: 5s
            - name: LOG_VERBOSE
              value: "false"
`,
	},
}
