package conn

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHClient(host, user, pwd string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		Timeout:         5 * time.Second,
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(pwd)},
	}
	addr := fmt.Sprintf("%s:%d", host, 22)
	return ssh.Dial("tcp", addr, sshConfig)
}
