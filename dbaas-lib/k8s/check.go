package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	v "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

type objects struct {
	Items []interface{} `json:"items"`
}

func (p *Cmd) PreCheck(name, version, operatorName, operatorImage, objectName string, supportedVersions map[string]string) ([]string, error) {
	warnings := []string{}
	deploymentImage, err := p.GetObjectsElement("deployment", operatorName, ".spec.template.spec.containers[0].image")
	if err != nil && err != ErrNotFound {
		return warnings, errors.Wrap(err, "get deployment image")
	}
	if err != nil && err == ErrNotFound {
		data, err := p.GetObjects(objectName)
		if err != nil && err == ErrNotFound {
			return warnings, nil
		} else if err != nil {
			return warnings, errors.Wrap(err, "get objects")
		}

		var obj objects
		err = json.Unmarshal(data, &obj)
		if err != nil {
			return warnings, err
		}
		if len(obj.Items) == 0 {
			return warnings, nil
		}

		return warnings, errors.Errorf("no operator but existing %s objects", objectName)
	}

	if len(deploymentImage) == 0 {
		return warnings, nil
	}

	if string(deploymentImage) == operatorImage {
		return warnings, nil
	}

	deployedVersion, err := getOperatorImageVersion(string(deploymentImage))
	if err != nil {
		return warnings, errors.Wrap(err, "get deployed operator image version")
	}

	operatorVersion, err := getOperatorImageVersion(operatorImage)
	if err != nil {
		return warnings, errors.Wrap(err, "get operator image version")
	}

	if _, ok := supportedVersions[deployedVersion]; !ok {
		return warnings, errors.New("not supported version " + deployedVersion)
	}
	dVersion, err := v.NewVersion(deployedVersion)
	if err != nil {
		return warnings, errors.New("convert deployed version " + deployedVersion)
	}
	oVersion, err := v.NewVersion(operatorVersion)
	if err != nil {
		return warnings, errors.New("convert version " + operatorVersion)
	}

	if oVersion.Compare(dVersion) != 0 {
		warnings = append(warnings, fmt.Sprintf("%s operator with version %s already installed. Trying work with it", objectName, deployedVersion))
	}

	return warnings, nil
}

func getOperatorImageVersion(image string) (string, error) {
	imageArr := strings.Split(image, ":")
	if len(imageArr) < 2 {
		return "", errors.New("no image version tag")
	}

	return imageArr[1], nil
}
