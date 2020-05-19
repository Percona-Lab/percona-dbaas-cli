package pxc

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	mrand "math/rand"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(name, opts, rootPass, version string) error {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return errors.Wrap(err, "version check")
	}
	err = p.ParseOptions(opts)
	if err != nil {
		return errors.Wrap(err, "parsing options")
	}

	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)
	switch p.platformType {
	case k8s.PlatformMinishift, k8s.PlatformMinikube:
		p.conf.SetupMiniConfig()
	}

	if len(rootPass) > 0 {
		err = p.SetupPasswords(name, rootPass)
		if err != nil {
			return errors.Wrap(err, "set root password")
		}
	}

	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}
	_, err = p.cmd.GetObjectsElement("deployment", p.operatorName(), ".spec.template.spec.containers[0].image")
	if err != nil && err == k8s.ErrNotFound {
		err = p.cmd.ApplyBundles(p.bundle)
		if err != nil {
			return errors.Wrap(err, "apply bundles")
		}
	}

	err = p.cmd.CreateCluster("pxc", p.conf.GetOperatorImage(), name, cr, p.bundle)
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// DeleteDBCluster delete cluster by name
func (p *PXC) DeleteDBCluster(name, opts, version string, delePVC bool) (string, error) {
	ext, err := p.cmd.IsObjExists("pxc", name)
	if err != nil {
		return "", errors.Wrap(err, "check if cluster exists")
	}

	if !ext {
		return "", errors.New("unable to find cluster pxc/" + name)
	}

	err = p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return "", errors.Wrap(err, "version check")
	}

	p.conf.SetName(name)

	err = p.cmd.DeleteCluster("pxc", p.operatorName(), name, delePVC)
	if err != nil {
		return "", errors.Wrap(err, "delete cluster")
	}
	if !delePVC {
		pvcObj, err := p.cmd.GetObject("pvc", "datadir-"+name+"-pxc-0")
		if err != nil {
			return "", errors.Wrap(err, "get pvc")
		}
		pvc := &corev1.PersistentVolumeClaim{}
		err = json.Unmarshal(pvcObj, pvc)
		if err != nil {
			return "", errors.Wrap(err, "unmarshal pvc")
		}
		return "pvc/" + pvc.Name, nil
	}
	err = p.cmd.DeleteObject("secret", name+"-secrets")
	if err != nil {
		return "", errors.Wrap(err, "delete secret")
	}

	return "", nil
}

// GetDBCluster return DB object
func (p *PXC) GetDBCluster(name, opts string) (dbaas.DB, error) {
	var db dbaas.DB
	err := p.setVersionObjectsWithDefaults(Version(""))
	if err != nil {
		return db, errors.Wrap(err, "version check")
	}
	secrets, err := p.cmd.GetSecrets(name + "-secrets")
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	cluster, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster object")

	}

	st := p.conf
	err = json.Unmarshal(cluster, st)
	if err != nil {
		return db, errors.Wrap(err, "unmarshal object")
	}
	err = p.checkClusterPods(name)
	if err != nil {
		db.Status = "error"
		return db, err
	}
	ns, err := p.cmd.GetCurrentNamespace()
	if err != nil {
		return db, errors.Wrap(err, "get namspace name")
	}
	if len(ns) == 0 {
		ns = "default"
	}
	db.Provider = provider
	db.Engine = engine
	db.ResourceName = name
	db.Port = 3306
	db.User = "root"
	db.Pass = string(secrets["root"])
	db.ResourceEndpoint = st.GetStatusHost() + "." + ns + ".pxc.svc.local"
	if p.conf.GetProxysqlServiceType() == "LoadBalancer" {
		svc := corev1.Service{}
		svcData, err := p.cmd.GetObject("svc", name+"-proxysql")
		if err != nil {
			return db, errors.Wrap(err, "get proxysql service")
		}
		err = json.Unmarshal(svcData, &svc)
		if err != nil {
			return db, errors.Wrap(err, "unmarshal proxysql service data")
		}
		for _, i := range svc.Status.LoadBalancer.Ingress {
			db.ResourceEndpoint = i.IP
			if len(i.Hostname) > 0 {
				db.ResourceEndpoint = i.Hostname
			}
		}
		if st.GetStatus() == dbaas.StateReady {
			db.Message = "To access database please run the following command:\nmysql -h " + db.ResourceEndpoint + " -P 3306 -uroot -pPASSWORD"
		}
		return db, nil
	}
	db.Status = st.GetStatus()
	if st.GetStatus() == dbaas.StateReady {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-proxysql 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
	}
	if st.GetStatus() == dbaas.StateUnknown && st.GetPXCStatus() == string(dbaas.StateReady) {
		db.Status = dbaas.StateReady
		db.Message = "To access database please run the following commands:\nkubectl port-forward pod/" + name + "-pxc-0 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
	}

	return db, nil
}

// GetDBClusterList return list of existing DB obkects
func (p *PXC) GetDBClusterList() ([]dbaas.DB, error) {
	var dbList []dbaas.DB
	cluster, err := p.cmd.GetObjects("pxc")
	if err != nil {
		return dbList, errors.Wrap(err, "get cluster object")

	}
	st := k8s.Clusters{}
	err = json.Unmarshal(cluster, &st)
	if err != nil {
		return dbList, errors.Wrap(err, "unmarshal object")
	}
	err = p.setVersionObjectsWithDefaults(Version(""))
	if err != nil {
		return dbList, errors.Wrap(err, "version check")
	}
	for _, c := range st.Items {
		b, err := json.Marshal(c)
		if err != nil {
			return dbList, errors.Wrap(err, "marshal")
		}

		err = json.Unmarshal(b, &p.conf)
		if err != nil {
			return dbList, errors.Wrap(err, "unmarshal pxc object")
		}
		pxc := p.conf
		db := dbaas.DB{
			ResourceName: pxc.GetName(),
			Status:       pxc.GetStatus(),
		}
		dbList = append(dbList, db)
	}

	return dbList, nil
}

// UpdateDBCluster update DB
func (p *PXC) UpdateDBCluster(name, opts, version string) error {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return errors.Wrap(err, "version check")
	}

	oldCR, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return errors.Wrap(err, "get cluster cr")
	}
	err = json.Unmarshal(oldCR, &p.conf)
	if err != nil {
		return errors.Wrap(err, "unmarshal cr")
	}

	err = p.ParseOptions(opts)
	if err != nil {
		return errors.Wrap(err, "parse options")
	}
	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)

	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.Upgrade("pxc", name, cr)
	if err != nil {
		return errors.Wrap(err, "upgrade cluster")
	}

	return nil
}

func (p *PXC) SetupPasswords(clusterName, rootPass string) error {
	secretName := clusterName + "-secrets"
	ext, err := p.cmd.IsObjExists("secret", secretName)
	if err != nil {
		return errors.Wrap(err, "check if secrets exists")
	}
	data := map[string][]byte{}
	if ext {
		data, err = p.cmd.GetSecrets(secretName)
		if err != nil {
			return errors.Wrap(err, "get secrets")
		}
		for k := range data {
			if k == "root" {
				data[k] = []byte(rootPass)
			}
		}
		err = p.cmd.UpdateSecrets(secretName, data)
		if err != nil {
			return errors.Wrap(err, "update secrets")
		}

		return nil
	}

	data["root"] = []byte(rootPass)
	data["xtrabackup"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create xtrabackup users password")
	}
	data["monitor"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create monitor users password")
	}
	data["clustercheck"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create clustercheck users password")
	}
	data["proxyadmin"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create proxyadmin users password")
	}

	err = p.cmd.CreateSecret(secretName, data)
	if err != nil {
		return errors.Wrap(err, "create secrets")
	}

	return nil
}

const (
	passwordMaxLen = 20
	passwordMinLen = 16
	passSymbols    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789"
)

func generatePass() ([]byte, error) {
	mrand.Seed(time.Now().UnixNano())
	ln := mrand.Intn(passwordMaxLen-passwordMinLen) + passwordMinLen
	b := make([]byte, ln)
	for i := 0; i < ln; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(passSymbols))))
		if err != nil {
			return nil, err
		}
		b[i] = passSymbols[randInt.Int64()]
	}

	return b, nil
}

func (p *PXC) PreCheck(name, opts, version string) ([]string, error) {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return nil, errors.Wrap(err, "version check")
	}
	supportedVersions := make(map[string]string)
	for v, obj := range objects {
		supportedVersions[string(v)] = obj.pxc.GetOperatorImage()
	}

	return p.cmd.PreCheck(name, string(version), p.operatorName(), p.conf.GetOperatorImage(), "pxc", supportedVersions)
}

func getOperatorImageVersion(image string) (string, error) {
	imageArr := strings.Split(image, ":")
	if len(imageArr) < 2 {
		return "", errors.New("no image version tag")
	}

	return imageArr[1], nil
}

func (p *PXC) checkClusterPods(name string) error {
	podsData, err := p.cmd.GetObjectByLables("pods", "app.kubernetes.io/instance="+name+",app.kubernetes.io/component=pxc")
	if err != nil {
		return errors.Wrap(err, "get pods")
	}
	err = checkPodsCondition(podsData)
	if err != nil {
		return err
	}

	podsData, err = p.cmd.GetObjectByLables("pods", "app.kubernetes.io/instance=cluster1,app.kubernetes.io/component=proxysql")
	if err != nil {
		return errors.Wrap(err, "get pods")
	}

	err = checkPodsCondition(podsData)
	if err != nil {
		return err
	}

	return nil
}

func checkPodsCondition(podsData []byte) error {
	var pods k8s.Pods
	err := json.Unmarshal(podsData, &pods)
	if err != nil {
		return errors.Wrap(err, "unmarshal pods data")
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase != "Pending" {
			return nil
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Status == "False" && strings.Contains(condition.Message, "Insufficient memory") {
				return k8s.ErrOutOfMemory
			}
		}
	}

	return nil
}
