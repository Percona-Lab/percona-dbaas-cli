package pxc

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
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
		pvc := &k8sPVC{}
		err = json.Unmarshal(pvcObj, pvc)
		if err != nil {
			return "", errors.Wrap(err, "unmarshal pvc")
		}
		return "pvc/" + pvc.Meta.Name, nil
	}
	err = p.cmd.DeleteObject("secret", name+"-secrets")
	if err != nil {
		return "", errors.Wrap(err, "delete secret")
	}

	return "", nil
}

// GetDBCluster return DB object
func (p *PXC) GetDBCluster(name, opts string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name + "-secrets")
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	cluster, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster object")

	}

	st := &k8sStatus{}
	err = json.Unmarshal(cluster, st)
	if err != nil {
		return db, errors.Wrap(err, "unmarshal object")
	}
	err = p.checkClusterPods(name, st)
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
	db.Status = string(st.Status.Status)
	db.ResourceEndpoint = st.Status.Host + "." + ns + ".pxc.svc.local"
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
		if st.Status.Status == "ready" {
			db.Message = "To access database please run the following command:\nmysql -h " + db.ResourceEndpoint + " -P 3306 -uroot -pPASSWORD"
		}
		return db, nil
	}

	if st.Status.Status == "ready" {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-proxysql 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
	}
	if st.Status.Status == "unknown" && st.Status.PXC.Status == "ready" {
		db.Status = "ready"
		db.Message = "To access database please run the following commands:\nkubectl port-forward pod/" + name + "-pxc-0 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
	}
	return db, nil
}

// GetDBClusterList return list of existing DB obkects
func (p *PXC) GetDBClusterList() ([]structs.DB, error) {
	var dbList []structs.DB
	cluster, err := p.cmd.GetObjects("pxc")
	if err != nil {
		return dbList, errors.Wrap(err, "get cluster object")

	}
	st := &k8sCluster{}
	err = json.Unmarshal(cluster, st)
	if err != nil {
		return dbList, errors.Wrap(err, "unmarshal object")
	}
	for _, c := range st.Items {
		db := structs.DB{
			ResourceName: c.Meta.Name,
			Status:       string(c.Status.Status),
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

func (p *PXC) checkClusterPods(name string, st *k8sStatus) error {
	for i := 0; i < int(st.Status.PXC.Size); i++ {
		podData, err := p.cmd.GetObject("pod", name+"-pxc-"+strconv.Itoa(i))
		if err != nil && err != k8s.ErrNotFound {
			return errors.Wrap(err, "get pxc pod data")
		} else if err != nil && err == k8s.ErrNotFound {
			continue
		}
		err = checkPodCondition(podData)
		if err != nil {
			return err
		}
	}
	for i := 0; i < int(st.Status.ProxySQL.Size); i++ {
		podData, err := p.cmd.GetObject("pod", name+"-proxysql-"+strconv.Itoa(i))
		if err != nil && err != k8s.ErrNotFound {
			return errors.Wrap(err, "get proxysql pod data")
		} else if err != nil && err == k8s.ErrNotFound {
			continue
		}
		err = checkPodCondition(podData)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkPodCondition(podData []byte) error {
	pod := k8s.Pod{}
	err := json.Unmarshal(podData, &pod)
	if err != nil {
		return errors.Wrap(err, "unmarshal pod data")
	}
	if pod.Status.Phase != "Pending" {
		return nil
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Status == "False" && strings.Contains(condition.Message, "Insufficient memory") {
			return k8s.ErrOutOfMemory
		}
	}
	return nil
}
