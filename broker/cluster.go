package broker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"
)

func (p *Controller) DeployPXCCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}

	dbservice.Namespace = instance.Context.Namespace

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	app, err := pxc.New(instance.Parameters.ClusterName, defaultVersion, "", p.EnvName)
	if err != nil {
		return err
	}
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

	app.ClusterConfig = conf

	var s3stor *dbaas.BackupStorageSpec

	setupmsg, err := app.Setup(conf, s3stor)
	if err != nil {
		log.Println("[Error] set configuration:", err)
		return nil
	}
	setupFinalMsg, err := SprintResponse("json", setupmsg)
	if err != nil {
		log.Println("[Error] sprint setup message", err)
	}
	log.Println(setupFinalMsg)

	created := make(chan pxc.ClusterData)
	msg := make(chan pxc.ClusterData)
	cerr := make(chan error)
	go app.Create(created, msg, cerr)
	go p.listenPXCCreateChannels(created, msg, cerr, instanceID, "pxc", dbservice)

	return nil
}

func (p *Controller) listenPXCCreateChannels(created chan pxc.ClusterData, msg chan pxc.ClusterData, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				p.instanceMap[instanceID].Credentials.Host = okmsg.Host
				p.instanceMap[instanceID].Credentials.Port = okmsg.Port
				p.instanceMap[instanceID].Credentials.Users = map[string]string{
					okmsg.User: okmsg.Pass,
				}
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal instance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			okFinalMsg, err := SprintResponse("json", okmsg)
			if err != nil {
				log.Println("[Error] sprint setup message", err)
			}
			log.Println(okFinalMsg)
			return
		case omsg := <-msg:
			log.Printf("[operator log error] %s\n", omsg)
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

func (p *Controller) DeployPSMDBCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}

	dbservice.Namespace = instance.Context.Namespace

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	app, err := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "", p.EnvName)
	if err != nil {
		return err
	}
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
	setupFinalMsg, err := SprintResponse("json", setupmsg)
	if err != nil {
		log.Println("[Error] sprint setup message", err)
	}
	log.Println(setupFinalMsg)

	created := make(chan psmdb.ClusterData)
	msg := make(chan psmdb.ClusterData)
	cerr := make(chan error)
	go app.Create(created, msg, cerr)
	go p.listenPSMDBCreateChannels(created, msg, cerr, instanceID, "psmdb", dbservice)

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

func (p *Controller) listenPSMDBCreateChannels(created chan psmdb.ClusterData, msg chan psmdb.ClusterData, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				p.instanceMap[instanceID].Credentials.Host = okmsg.Host
				p.instanceMap[instanceID].Credentials.Port = okmsg.Port
				p.instanceMap[instanceID].Credentials.Users = map[string]string{
					okmsg.UserAdminUser:    okmsg.UserAdminPass,
					okmsg.ClusterAdminUser: okmsg.ClusterAdminPass,
				}
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal instance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			okFinalMsg, err := SprintResponse("json", okmsg)
			if err != nil {
				log.Println("[Error] sprint setup message", err)
			}
			log.Println(okFinalMsg)
			return
		case omsg := <-msg:
			log.Println("[operator log error] ", omsg.Message)
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
		app, err := pxc.New(name, defaultVersion, "", p.EnvName)
		if err != nil {
			return err
		}
		go app.Delete(delePVC, ok, cerr)
		p.listenDeleteChannels(ok, cerr)
	case psmdbServiceID:
		app, err := psmdb.New(name, name, defaultVersion, "", p.EnvName)
		if err != nil {
			return err
		}
		go app.Delete(delePVC, ok, cerr)
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

func (p *Controller) UpdatePXCCluster(instance *ServiceInstance) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}
	created := make(chan pxc.ClusterData)
	msg := make(chan pxc.ClusterData)
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

	app, err := pxc.New(instance.Parameters.ClusterName, defaultVersion, "", p.EnvName)
	if err != nil {
		return err
	}
	conf := pxc.ClusterConfig{}
	SetPXCDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PXC.Instances = instance.Parameters.Replicas
	}

	conf.PXC.BrokerInstance = string(brokerInstance)
	app.ClusterConfig = conf

	go app.Edit(nil, created, msg, cerr)
	p.listenPXCUpdateChannels(created, msg, cerr, instance.ID, "pxc", p.dbaas)

	return nil
}

func (p *Controller) listenPXCUpdateChannels(created chan pxc.ClusterData, msg chan pxc.ClusterData, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
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
			log.Println(okmsg)
			return
		case omsg := <-msg:
			log.Println("[operator log error] \n", omsg)
		case err := <-cerr:
			p.instanceMap[instanceID].LastOperation.State = FailedOperationState
			p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			log.Println("Create error:", err)
			return
		}
	}
}

func (p *Controller) UpdatePSMDBCluster(instance *ServiceInstance) error {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return err
	}
	created := make(chan psmdb.ClusterData)
	msg := make(chan psmdb.ClusterData)
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

	app, err := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "", p.EnvName)
	if err != nil {
		return err
	}
	conf := psmdb.ClusterConfig{}
	SetPSMDBDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PSMDB.Instances = instance.Parameters.Replicas
	}
	conf.PSMDB.BrokerInstance = string(brokerInstance)
	app.ClusterConfig = conf

	go app.Edit(nil, created, msg, cerr)
	p.listenPSMDBUpdateChannels(created, msg, cerr, instance.ID, "psmdb", p.dbaas)

	return nil
}

func (p *Controller) listenPSMDBUpdateChannels(created chan psmdb.ClusterData, msg chan psmdb.ClusterData, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
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
			log.Println(okmsg)
			return
		case omsg := <-msg:
			log.Printf("[operator log error] %s\n", omsg)
		case err := <-cerr:
			p.instanceMap[instanceID].LastOperation.State = FailedOperationState
			p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			log.Println("Create error:", err)
			return
		}
	}
}

func SprintResponse(output string, data interface{}) (string, error) {
	if output == "json" {
		d, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		return fmt.Sprintln(string(d)), nil
	}

	return fmt.Sprintln(data), nil
}
