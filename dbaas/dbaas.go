package dbaas

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type ErrCmdRun struct {
	cmd    string
	args   []string
	output []byte
}

func (e ErrCmdRun) Error() string {
	return fmt.Sprintf("failed to run `%s %s`, output: %s", e.cmd, strings.Join(e.args, " "), e.output)
}

func runCmd(cmd string, args ...string) ([]byte, error) {
	o, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return nil, ErrCmdRun{cmd: cmd, args: args, output: o}
	}

	return o, nil
}

func readOperatorLogs(operatorName string) ([]byte, error) {
	return runCmd("kubectl", "logs", "-l", "name="+operatorName)
}

func getCR(typ, clusterName string) ([]byte, error) {
	return runCmd("kubectl", "get", typ+"/"+clusterName, "-o", "json")
}

func apply(k8sObj string) error {
	_, err := runCmd("sh", "-c", "cat <<-EOF | kubectl apply -f -\n"+k8sObj+"\nEOF")
	if err != nil {
		return err
	}

	return nil
}

func IsCRexists(typ, name string) (bool, error) {
	switch typ {
	case "pxc":
		typ = "perconaxtradbcluster.pxc.percona.com"
	}

	out, err := runCmd("kubectl", "get", typ, name, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return false, errors.Wrapf(err, "get cr: %s", out)
	}

	return strings.TrimSpace(string(out)) == typ+"/"+name, nil
}
