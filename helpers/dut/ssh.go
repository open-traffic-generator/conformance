package dut

import (
	"bytes"
	"fmt"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

type SshClient struct {
	location string
	username string
	password string
	config   *ssh.ClientConfig
	client   *ssh.Client
}

func NewSshClient(t *testing.T, location string, username string, password string) (*SshClient, error) {
	t.Logf("Creating SSH client for server %s ...\n", location)

	authMethod := []ssh.AuthMethod{}

	if password != "" {
		authMethod = append(authMethod, ssh.Password(password))
	} else {
		authMethod = append(authMethod, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			return nil, nil
		}))
	}

	sshConfig := ssh.ClientConfig{
		User: username,
		Auth: authMethod,
		HostKeyCallback: ssh.HostKeyCallback(
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// TODO: replace this with a more appropriate validation
				if len(key.Marshal()) == 0 {
					return fmt.Errorf("key is empty")
				}
				return nil
			}),
	}

	t.Logf("Dialing ssh://%s@%s ...\n", username, location)

	client, err := ssh.Dial("tcp", location, &sshConfig)
	if err != nil {
		return nil, fmt.Errorf("could not dial ssh://%s@%s: %v", username, location, err)
	}

	return &SshClient{
		location: location,
		username: username,
		password: password,
		config:   &sshConfig,
		client:   client,
	}, nil
}

func CloseSshClient(t *testing.T, c *SshClient) {
	t.Logf("Closing SSH client for server %s ...\n", c.location)
	c.client.Close()
}

func ExecSshCmd(t *testing.T, c *SshClient, cmd string, logIO bool) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not create ssh session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if logIO {
		t.Logf("SSH INPUT: \n%s\n", cmd)
	}
	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("could not execute '%s': %v", cmd, err)
	}
	out := b.String()

	if logIO {
		t.Logf("SSH OUTPUT: \n%s\n", out)
	}
	return out, nil
}
