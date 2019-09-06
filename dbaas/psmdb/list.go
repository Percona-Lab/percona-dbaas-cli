package psmdb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

const describeMsg = `
Name:                                          %v
Status:                                        %v
Multi-AZ:                                      %v
Labels:                                        %v
 
PSMDB Count:                                   %v
PSMDB Image:                                   %v
PSMDB CPU Requests:                            %v
PSMDB Memory Requests:                         %v
PSMDB PodDisruptionBudget:                     %v
PSMDB AntiAffinityTopologyKey:                 %v
PSMDB StorageType:                             %v
PSMDB Allocated Storage:                       %v
 
Backup coordinator Image:                      %v
Backup coordinator CPU Requests:               %v
Backup coordinator Memory Requests:            %v
Backup coordinator AntiAffinityTopologyKey:    %v
Backup coordinator StorageType:                %v
Backup coordinator Allocated Storage:          %v
 
Backup Schedule:                               %v
`

func (p *PSMDB) describe(kubeInput []byte) (string, error) {
	cr := &PerconaServerMongoDB{}
	err := json.Unmarshal([]byte(kubeInput), &cr)
	if err != nil {
		return "", errors.Wrapf(err, "json prase")
	}

	multiAz := "yes"
	noAzAffinityList := []string{"none", "hostname"}
	for _, arg := range noAzAffinityList {
		if *cr.Spec.Replsets[0].Affinity.TopologyKey == arg {
			multiAz = "no"
		}
	}
	budgetPSMDB := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}

	if cr.Spec.Replsets[0].PodDisruptionBudget != nil && cr.Spec.Replsets[0].PodDisruptionBudget.MinAvailable != nil {
		budgetPSMDB["MinAvailable"] = cr.Spec.Replsets[0].PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.Replsets[0].PodDisruptionBudget != nil && cr.Spec.Replsets[0].PodDisruptionBudget.MaxUnavailable != nil {
		budgetPSMDB["MaxUnavailable"] = cr.Spec.Replsets[0].PodDisruptionBudget.MaxUnavailable.String()
	}
	budgetCoordinator := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}
	if cr.Spec.Backup.Coordinator.PodDisruptionBudget != nil && cr.Spec.Backup.Coordinator.PodDisruptionBudget.MinAvailable != nil {
		budgetCoordinator["MinAvailable"] = cr.Spec.Backup.Coordinator.PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.Backup.Coordinator.PodDisruptionBudget != nil && cr.Spec.Backup.Coordinator.PodDisruptionBudget.MaxUnavailable != nil {
		budgetCoordinator["MaxUnavailable"] = cr.Spec.Backup.Coordinator.PodDisruptionBudget.MaxUnavailable.String()
	}

	cpuSizeBytes, err := cr.Spec.Backup.Coordinator.Resources.Requests["cpu"].MarshalJSON()
	if err != nil {
		return "", err
	}
	memorySizeBytes, err := cr.Spec.Backup.Coordinator.Resources.Requests["memory"].MarshalJSON()
	if err != nil {
		return "", err
	}

	backupAffinity := "not set"
	backupSchedule := "not set"
	if cr.Spec.Backup.Coordinator.Affinity != nil && cr.Spec.Backup.Coordinator.Affinity.TopologyKey != nil {
		backupAffinity = *cr.Spec.Backup.Coordinator.Affinity.TopologyKey
	}
	if cr.Spec.Backup.Tasks != nil {
		for index := range cr.Spec.Backup.Tasks {
			backupSchedule = backupSchedule + cr.Spec.Backup.Tasks[index].Name + ", "
		}
		backupSchedule = strings.TrimSuffix(backupSchedule, ", ")
	}

	return fmt.Sprintf(describeMsg,
		cr.ObjectMeta.Name,
		cr.Status.Status,
		multiAz,
		dbaas.GetStringFromMap(cr.ObjectMeta.Labels),
		cr.Spec.Replsets[0].Size,
		cr.Spec.Image,
		cr.Spec.Replsets[0].Resources.Requests.CPU,
		cr.Spec.Replsets[0].Resources.Requests.Memory,
		dbaas.GetStringFromMap(budgetPSMDB),
		*cr.Spec.Replsets[0].Affinity.TopologyKey,
		cr.StorageClassesAllocated.DataPod,
		cr.StorageSizeAllocated.DataPod,
		cr.Spec.Backup.Image,
		string(cpuSizeBytes),
		string(memorySizeBytes),
		backupAffinity,
		cr.StorageClassesAllocated.BackupCoordinator,
		cr.StorageSizeAllocated.BackupCoordinator,
		backupSchedule), nil
}

func (p *PSMDB) Describe() (string, error) {
	out, err := p.Cmd.RunCmd("kubectl", "get", p.typ, p.name, "-o", "json")
	if err != nil {
		return "", errors.Wrapf(err, "describe-db %s", out)
	}

	mergedData := map[string]interface{}{}
	err = json.Unmarshal([]byte(out), &mergedData)
	if err != nil {
		return "", errors.Wrapf(err, "describe-db")
	}

	pvcList := &dbaas.MultiplePVCk8sOutput{}
	pvcsJSON, err := p.Cmd.RunCmd("kubectl", "get", "pvc", fmt.Sprintf("--selector=app.kubernetes.io/instance=%s,app.kubernetes.io/managed-by=%s", p.name, p.OperatorName()), "-o", "json")
	if err != nil {
		return "", errors.Wrapf(err, "describe-db")
	}
	err = json.Unmarshal([]byte(pvcsJSON), pvcList)
	if err != nil {
		return "", errors.Wrapf(err, "describe-db")
	}
	PVCs := map[string]string{}
	AllocatedStorage := map[string]string{}

	for volume := range pvcList.Items {
		PVCs[pvcList.Items[volume].Labels["app.kubernetes.io/component"]] = *pvcList.Items[volume].Spec.StorageClassName
		qt, err := pvcList.Items[volume].Status.Capacity["storage"].MarshalJSON()
		if err != nil {
			return "", errors.Wrapf(err, "describe-db")
		}
		AllocatedStorage[pvcList.Items[volume].Labels["app.kubernetes.io/component"]] = string(qt)
	}
	mergedData["StorageClassesAllocated"] = PVCs
	mergedData["StorageSizeAllocated"] = AllocatedStorage

	out, err = json.Marshal(mergedData)
	if err != nil {
		return "", errors.Wrapf(err, "describe-db")
	}
	return p.describe(out)
}

func (p *PSMDB) List() (string, error) {
	return p.Cmd.List(p.typ)
}
