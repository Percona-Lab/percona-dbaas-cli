// Copyright © 2019 Percona, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbaas

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func (p Cmd) Upgrade(typ string, app Deploy, operator string, apps map[string]string, ok chan<- string, msg chan<- OutuputMsg, errc chan<- error) {
	if operator != "" {
		err := p.upgradeOperator(app, operator)
		if err != nil {
			errc <- errors.Wrap(err, "upgrade operator")
			return
		}
	}

	acr, err := p.GetObject(typ, app.Name())
	if err != nil {
		errc <- errors.Wrap(err, "get config")
		return
	}

	err = app.Upgrade(acr, apps)
	if err != nil {
		errc <- errors.Wrap(err, "upgrade cluster")
		return
	}

	cr, err := app.App()
	if err != nil {
		errc <- errors.Wrap(err, "get cr")
		return
	}
	err = p.apply(cr)
	if err != nil {
		errc <- errors.Wrap(err, "apply cr")
		return
	}

	time.Sleep(15 * time.Second)
	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := p.GetObject(typ, app.Name())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := app.CheckStatus(status, make(map[string][]byte))
		if err != nil {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}

		switch state {
		case ClusterStateReady:
			ok <- strings.Join(msgs, "\n")
			return
		case ClusterStateError:
			errc <- errors.New(strings.Join(msgs, "\n"))
			return
		case ClusterStateInit:
		}

		opLogsStream, err := p.readOperatorLogs(app.OperatorName())
		if err != nil {
			errc <- errors.Wrap(err, "get operator logs")
			return
		}

		opLogs, err := app.CheckOperatorLogs(opLogsStream)
		if err != nil {
			errc <- errors.Wrap(err, "parse operator logs")
			return
		}

		for _, entry := range opLogs {
			msg <- entry
		}

		if tries >= getStatusMaxTries {
			errc <- errors.Wrap(err, "unable to start cluster")
			return
		}

		tries++
	}
}

func (p Cmd) upgradeOperator(app Deploy, newImage string) error {
	if newImage == "" {
		return nil
	}

	for _, o := range app.Bundle(newImage) {
		if o.Kind == "Deployment" && o.Name == app.OperatorName() {
			err := p.apply(o.Data)
			if err != nil {
				return errors.Wrap(err, "apply cr")
			}

			time.Sleep(15 * time.Second)
			tries := 0
			tckr := time.NewTicker(500 * time.Millisecond)
			defer tckr.Stop()
			for range tckr.C {
				status, err := p.runCmd(execCommand, "get", "pod", "-l", "name="+app.OperatorName(), "-o", "json")
				if err != nil {
					return errors.Wrap(err, "get status")
				}
				pods := corev1.PodList{}
				err = json.Unmarshal(status, &pods)
				if err != nil {
					return errors.Wrap(err, "marshal status")
				}

				if len(pods.Items) < 1 {
					return errors.Wrapf(err, "unable to find operator pod for %s", app.OperatorName())
				}

				pod := pods.Items[0]
				switch pod.Status.Phase {
				case corev1.PodRunning:
					return nil
				case corev1.PodFailed:
					return errors.Errorf("failed to run: %s: %s", pod.Status.Message, pod.Status.Reason)
				default:
				}

				if tries >= getStatusMaxTries {
					return errors.Wrap(err, "response limit has reached, unable get the success status from pod")
				}

				tries++
			}
		}
	}
	return nil
}
