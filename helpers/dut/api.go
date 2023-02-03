package dut

import (
	"fmt"
	"strings"
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/testconfig"
)

type DutApi struct {
	t         *testing.T
	dutConfig *testconfig.DutConfig
	sshClient *SshClient
}

func NewDutApi(t *testing.T, dc *testconfig.DutConfig) *DutApi {
	t.Logf("DUT Host: %s\n", dc.Host)
	t.Logf("DUT Interfaces: %s\n", dc.Interfaces)
	t.Logf("DUT SSH Port: %v\n", dc.SshPort)
	t.Logf("DUT gNMI Port: %v\n", dc.GnmiPort)

	return &DutApi{
		t:         t,
		dutConfig: dc,
	}
}

func (d *DutApi) ExecSshCmd(cmd string) (string, error) {
	dc := d.dutConfig
	if d.sshClient == nil {
		c, err := NewSshClient(
			d.t, fmt.Sprintf("%s:%d", dc.Host, dc.SshPort),
			dc.SshUsername, dc.SshPassword,
		)
		if err != nil {
			return "", fmt.Errorf("could not create SSH client: %v", err)
		}

		d.sshClient = c
	}

	return ExecSshCmd(d.t, d.sshClient, cmd, true)
}

func (d *DutApi) PushSshConfig(cfg string) (string, error) {
	lines := strings.Split(cfg, "\n")
	newLines := make([]string, 0, len(lines)+2)
	newLines = append(newLines, "enable", "config terminal")
	for _, line := range lines {
		line = strings.TrimPrefix(line, "		")
		if len(line) != 0 {
			newLines = append(newLines, line)
		}
	}

	return d.ExecSshCmd(strings.Join(newLines, "\n") + "\n")
}

func (d *DutApi) SetSshConfig(setCfg string, unsetCfg string) func() {
	d.t.Log("Setting SSH config on DUT ...")
	if _, err := d.PushSshConfig(setCfg); err != nil {
		d.t.Error("Failure occurred while pushing SSH config, undoing it ...")
		if _, e := d.PushSshConfig(unsetCfg); e != nil {
			d.t.Errorf("Failure occurred while undoing SSH config: %v\n", e)
		}
		d.t.Fatal(err)
	}

	return func() {
		d.t.Log("Un-setting SSH config on DUT ...")
		if _, err := d.PushSshConfig(unsetCfg); err != nil {
			d.t.Fatal(err)
		}
	}
}
