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

package psmdb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// PSMDBCmd represents the pxc command
var PSMDBCmd = &cobra.Command{
	Use:   "psmdb",
	Short: "Manage your MongoDB cluster on Kubernetes",
}

func parseArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	if a := strings.Split(args[0], "/"); len(a) == 2 {
		args = a
	}

	return args
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
