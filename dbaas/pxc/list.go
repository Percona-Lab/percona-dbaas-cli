package pxc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

const describeMsg = `
Name:                                %v
Status:                              %v
Multi-AZ:                            %v
Labels:                              %v
 
PXC Count:                           %v
PXC Image:                           %v
PXC CPU Requests:                    %v
PXC Memory Requests:                 %v
PXC PodDisruptionBudget:             %v
PXC AntiAffinityTopologyKey:         %v
PXC StorageType:                     %v
PXC Allocated Storage:               %v
 
ProxySQL Count:                      %v
ProxySQL Image:                      %v
ProxySQL CPU Requests:               %v
ProxySQL Memory Requests:            %v
ProxySQL PodDisruptionBudget:        %v
ProxySQL AntiAffinityTopologyKey:    %v
ProxySQL StorageType:                %v
ProxySQL Allocated Storage:          %v
 
Backup Image:                        %v
Backup StorageType:                  %v
Backup Allocated Storage:            %v
Backup Schedule:                     %v
`

func (p *PXC) describe(kubeInput []byte) (string, error) {
	cr := &PerconaXtraDBCluster{}
	err := json.Unmarshal([]byte(kubeInput), &cr)
	if err != nil {
		return "", errors.Wrapf(err, "json prase")
	}

	multiAz := "yes"
	noAzAffinityList := []string{"none", "hostname"}
	for _, arg := range noAzAffinityList {
		if *cr.Spec.PXC.Affinity.TopologyKey == arg {
			multiAz = "no"
		}
	}
	budgetPXC := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}

	if cr.Spec.PXC.PodDisruptionBudget != nil && cr.Spec.PXC.PodDisruptionBudget.MinAvailable != nil {
		budgetPXC["MinAvailable"] = cr.Spec.PXC.PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.PXC.PodDisruptionBudget != nil && cr.Spec.PXC.PodDisruptionBudget.MaxUnavailable != nil {
		budgetPXC["MaxUnavailable"] = cr.Spec.PXC.PodDisruptionBudget.MaxUnavailable.String()
	}
	budgetSQL := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}
	if cr.Spec.ProxySQL.PodDisruptionBudget != nil && cr.Spec.ProxySQL.PodDisruptionBudget.MinAvailable != nil {
		budgetSQL["MinAvailable"] = cr.Spec.ProxySQL.PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.ProxySQL.PodDisruptionBudget != nil && cr.Spec.ProxySQL.PodDisruptionBudget.MaxUnavailable != nil {
		budgetSQL["MaxUnavailable"] = cr.Spec.ProxySQL.PodDisruptionBudget.MaxUnavailable.String()
	}

	backupImage := "not set"
	backupSize := "not set"
	backupStorageClassName := "not set"
	backupSchedule := "not set"
	if cr.Spec.Backup != nil {
		backupImage = cr.Spec.Backup.Image

		if cr.Spec.Backup.Schedule != nil {
			backupSchedule = ""
			for schedule := range cr.Spec.Backup.Schedule {
				backupSchedule = backupSchedule + cr.Spec.Backup.Schedule[schedule].Name + ", "
			}
		}
		backupSchedule = strings.TrimRight(backupSchedule, ", ")
		for item := range cr.Spec.Backup.Storages {
			if cr.Spec.Backup.Storages[item].Type == "filesystem" {
				volume := cr.Spec.Backup.Storages[item]
				backupSizeBytes, err := volume.Volume.PersistentVolumeClaim.Resources.Requests["storage"].MarshalJSON()
				if err != nil {
					return "", err
				}
				backupSize = string(backupSizeBytes)
				backupStorageClassName = string(*volume.Volume.PersistentVolumeClaim.StorageClassName)
			}

		}
	}

	return fmt.Sprintf(describeMsg,
		cr.ObjectMeta.Name,
		cr.Status.Status,
		multiAz,
		dbaas.GetStringFromMap(cr.ObjectMeta.Labels),
		cr.Spec.PXC.Size,
		cr.Spec.PXC.Image,
		cr.Spec.PXC.Resources.Requests.CPU,
		cr.Spec.PXC.Resources.Requests.Memory,
		dbaas.GetStringFromMap(budgetPXC),
		*cr.Spec.PXC.Affinity.TopologyKey,
		cr.StorageClassesAllocated.PXC,
		cr.StorageSizeAllocated.PXC,
		cr.Spec.ProxySQL.Size,
		cr.Spec.ProxySQL.Image,
		cr.Spec.ProxySQL.Resources.Requests.CPU,
		cr.Spec.ProxySQL.Resources.Requests.Memory,
		dbaas.GetStringFromMap(budgetSQL),
		*cr.Spec.ProxySQL.Affinity.TopologyKey,
		cr.StorageClassesAllocated.ProxySQL,
		cr.StorageSizeAllocated.ProxySQL,
		backupImage,
		backupSize,
		backupStorageClassName,
		backupSchedule), nil
}

func (p *PXC) Describe() (string, error) {
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

func (p *PXC) List() (string, error) {
	return p.Cmd.List(p.typ)
}
