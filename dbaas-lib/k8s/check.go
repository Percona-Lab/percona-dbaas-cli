package k8s

import (
	"fmt"
	"strings"

	v "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

func (p *Cmd) PreCheck(name, version, operatorName, operatorImage, objectName string, supportedVersions map[string]string) (warnings []string, errArr []error) {
	deploymentImage, err := p.GetObjectsElement("deployment", operatorName, ".spec.template.spec.containers[0].image")
	if err != nil && err != ErrNotFound {
		errArr = append(errArr, errors.Wrap(err, "get deployment image"))
		return
	}
	if err != nil && err == ErrNotFound {
		_, err = p.GetObjects(objectName)
		if err != nil && err == ErrNotFound {

			return
		} else if err != nil {

			errArr = append(errArr, err)
		}
	}

	if string(deploymentImage) == operatorImage {
		return
	}

	deployedVersion, err := getOperatorImageVersion(string(deploymentImage))
	if err != nil {
		errArr = append(errArr, err)
		return
	}

	operatorVersion, err := getOperatorImageVersion(operatorImage)
	if err != nil {
		errArr = append(errArr, err)
		return
	}

	if _, ok := supportedVersions[deployedVersion]; !ok {
		errArr = append(errArr, errors.New("not supported version "+deployedVersion))
		return
	}
	dVersion, err := v.NewVersion(deployedVersion)
	if err != nil {
		errArr = append(errArr, errors.New("convert version "+deployedVersion))
		return
	}
	oVersion, err := v.NewVersion(operatorVersion)
	if err != nil {
		errArr = append(errArr, errors.New("convert version "+operatorVersion))
		return
	}

	i := oVersion.Compare(dVersion)
	switch i {
	case 0:
		return
	default:
		warnings = append(warnings, fmt.Sprintf("%s operator with version %s already installed. Trying work with it", objectName, deployedVersion))
	}
	return
}

func getOperatorImageVersion(image string) (string, error) {
	imageArr := strings.Split(image, ":")
	if len(imageArr) < 2 {
		return "", errors.New("no image version tag")
	}

	return imageArr[1], nil
}
