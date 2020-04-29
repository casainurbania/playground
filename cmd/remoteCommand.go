package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type RemoteCommand struct {
	BaseCommand
	client *ssh.Client
}

func NewRemoteCmd(cmd string, timeout int, client *ssh.Client) (*RemoteCommand, error) {
	c := &RemoteCommand{}
	c.Cmd = cmd
	c.Timeout = time.Duration(timeout) * time.Second

	if c.Timeout <= 0*time.Second {
		c.Timeout = DEFAULT_RUM_TIMEOUT * time.Second
	}
	c.TerminateChan = make(chan int)
	c.client = client

	return c, nil
}

func (c *RemoteCommand) Run() error {
	//创建session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	errChan := make(chan error)
	resChan := make(chan []byte)
	go func() {
		resb, err := session.CombinedOutput(c.Cmd)
		if err != nil {
			errChan <- err
			return
		}
		resChan <- resb
	}()
	select {
	case err = <-errChan: // 执行命令出错
		c.stderr.Write([]byte(err.Error()))
	case res := <-resChan: // 执行命令正常返回
		c.stdout.Write(res)
	case <-time.After(c.Timeout): // 执行超时
		err = errors.New(fmt.Sprintf("cmd run timeout, cmd [%s], time[%v]", c.Cmd, c.Timeout))
	case <-c.TerminateChan: // 被中断
		err = errors.New(fmt.Sprintf("cmd is terminated, cmd [%s]", c.Cmd))
	}

	return err
}

func (c *RemoteCommand) Stderr() string {
	return strings.TrimSpace(string(c.stderr.Bytes()))
}

func (c *RemoteCommand) Stdout() string {
	return strings.TrimSpace(string(c.stdout.Bytes()))
}
