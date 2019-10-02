package broker

import (
	"encoding/json"
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"
)

func (p *Controller) DeployCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}

	dbservice.Namespace = instance.Context.Namespace

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return err
	}
	switch instance.ServiceID {
	case pxcServiceID:
		app := pxc.New(instance.Parameters.ClusterName, defaultVersion, "")
		conf := pxc.ClusterConfig{}
		SetPXCDefaults(&conf)
		if instance.Parameters.Replicas > int32(0) {
			conf.PXC.Instances = instance.Parameters.Replicas
		}
		if len(instance.Parameters.Size) > 0 {
			conf.PXC.StorageSize = instance.Parameters.Size
		}
		if len(instance.Parameters.TopologyKey) > 0 {
			conf.PXC.AntiAffinityKey = instance.Parameters.TopologyKey
		}
		conf.PXC.BrokerInstance = string(brokerInstance)

		if instance.Parameters.PMM.Enabled {
			conf.PMM.Image = instance.Parameters.PMM.Image
			conf.PMM.ServerHost = instance.Parameters.PMM.Host
			conf.PMM.ServerUser = instance.Parameters.PMM.User
			conf.PMM.ServerPass = instance.Parameters.PMM.Pass
			conf.PMM.Enabled = instance.Parameters.PMM.Enabled
		}

		app.ClusterConfig = conf

		var s3stor *dbaas.BackupStorageSpec

		setupmsg, err := app.Setup(conf, s3stor, dbservice.GetPlatformType())
		if err != nil {
			log.Println("[Error] set configuration:", err)
			return nil
		}

		log.Println(setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		go dbservice.Create("pxc", app, instance.Parameters.OperatorImage, created, msg, cerr)
		go p.listenCreateChannels(created, msg, cerr, instanceID, "pxc", dbservice)
	case psmdbServiceID:
		app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "")
		conf := psmdb.ClusterConfig{}
		SetPSMDBDefaults(&conf)
		if instance.Parameters.Replicas > int32(0) {
			conf.PSMDB.Instances = instance.Parameters.Replicas
		}
		if len(instance.Parameters.Size) > 0 {
			conf.PSMDB.StorageSize = instance.Parameters.Size
		}
		if len(instance.Parameters.TopologyKey) > 0 {
			conf.PSMDB.AntiAffinityKey = instance.Parameters.TopologyKey
		}
		conf.PSMDB.BrokerInstance = string(brokerInstance)
		app.ClusterConfig = conf

		var s3stor *dbaas.BackupStorageSpec

		setupmsg, err := app.Setup(s3stor, dbservice.GetPlatformType())
		if err != nil {
			log.Println("[Error] set configuration:", err)
			return nil
		}

		log.Println(setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		go dbservice.Create("psmdb", app, "", created, msg, cerr)
		go p.listenCreateChannels(created, msg, cerr, instanceID, "psmdb", dbservice)
	}
	return nil
}

func (p *Controller) getClusterSecret(clusterName string) (Secret, error) {
	var secret Secret
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return secret, err
	}

	s, err := dbservice.GetObject("secret", clusterName+"-secrets")
	if err != nil {
		return secret, err
	}
	err = json.Unmarshal(s, &secret)

	return secret, err
}

func (p *Controller) listenCreateChannels(created chan string, msg chan dbaas.OutuputMsg, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			var credentials Credentials
			err := json.Unmarshal([]byte(okmsg), &credentials)
			if err != nil {
				log.Println("Error unmarshal credentials:", err)
			}
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				p.instanceMap[instanceID].Credentials = credentials
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal instance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			log.Printf(okmsg)
			return
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				//log.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				log.Printf("[operator log error] %s\n", omsg)
			}
		case err := <-cerr:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = FailedOperationState
				p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			}
			log.Println("Create error:", err)
			return
		}
	}
}

func (p *Controller) DeleteCluster(instance *ServiceInstance) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := false
	name := instance.Parameters.ClusterName
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}

	dbservice.Namespace = instance.Context.Namespace

	switch instance.ServiceID {
	case pxcServiceID:
		go dbservice.Delete("pxc", pxc.New(name, defaultVersion, ""), delePVC, ok, cerr)
		p.listenDeleteChannels(ok, cerr)
	case psmdbServiceID:
		go dbservice.Delete("psmdb", psmdb.New(name, name, defaultVersion, ""), delePVC, ok, cerr)
		p.listenDeleteChannels(ok, cerr)
	}
	return nil
}

func (p *Controller) listenDeleteChannels(ok chan string, cerr chan error) {
	for {
		select {
		case <-ok:
			log.Println("Deleting...[done]")
			return
		case err := <-cerr:
			log.Printf("[ERROR] delete pxc: %v", err)
			return
		}
	}
}

func (p *Controller) UpdateCluster(instance *ServiceInstance) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}
	created := make(chan string)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)
	dbservice, err = dbaas.New(p.EnvName)
	if err != nil {
		return err
	}

	dbservice.Namespace = instance.Context.Namespace

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return err
	}
	switch instance.ServiceID {
	case pxcServiceID:
		app := pxc.New(instance.Parameters.ClusterName, defaultVersion, "")
		conf := pxc.ClusterConfig{}
		SetPXCDefaults(&conf)
		if instance.Parameters.Replicas > int32(0) {
			conf.PXC.Instances = instance.Parameters.Replicas
		}

		conf.PXC.BrokerInstance = string(brokerInstance)
		app.ClusterConfig = conf

		go dbservice.Edit("pxc", app, nil, created, msg, cerr)
		p.listenUpdateChannels(created, msg, cerr, instance.ID, "pxc", p.dbaas)
	case psmdbServiceID:
		app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "")
		conf := psmdb.ClusterConfig{}
		SetPSMDBDefaults(&conf)
		if instance.Parameters.Replicas > int32(0) {
			conf.PSMDB.Instances = instance.Parameters.Replicas
		}
		conf.PSMDB.BrokerInstance = string(brokerInstance)
		app.ClusterConfig = conf

		go dbservice.Edit("psmdb", app, nil, created, msg, cerr)
		p.listenUpdateChannels(created, msg, cerr, instance.ID, "psmdb", p.dbaas)
	}
	return nil
}

func (p *Controller) listenUpdateChannels(created chan string, msg chan dbaas.OutuputMsg, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal oinstance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			log.Printf(okmsg)
			return
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				// fmt.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				log.Printf("[operator log error] %s\n", omsg)
			}
		case err := <-cerr:
			p.instanceMap[instanceID].LastOperation.State = FailedOperationState
			p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			log.Println("Create error:", err)
			return
		}
	}
}
