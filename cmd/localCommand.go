package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	DEFAULT_RUM_TIMEOUT = 3600
)

type LocalCommand struct {
	BaseCommand
	Setpgid      bool
	LocalCommand *exec.Cmd
}

func NewLocalCmd(c *LocalCommand) (*LocalCommand, error) {
	if c.Timeout == 0*time.Second {
		c.Timeout = DEFAULT_RUM_TIMEOUT * time.Second
	}
	if c.TerminateChan == nil {
		c.TerminateChan = make(chan int)
	}
	cmd := exec.Command("/bin/bash", "-c", c.Cmd)
	if c.Setpgid {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	cmd.Stderr = &c.stderr
	cmd.Stdout = &c.stdout
	c.LocalCommand = cmd

	return c, nil
}

func (c *LocalCommand) Run() error {
	if err := c.LocalCommand.Start(); err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		errChan <- c.LocalCommand.Wait()
		defer close(errChan)
	}()

	var err error
	select {
	case err = <-errChan:
	case <-time.After(c.Timeout):
		err = c.terminate()
		if err == nil {
			err = errors.New(fmt.Sprintf("cmd run timeout, cmd [%s], time[%v]", c.Cmd, c.Timeout))
		}
	case <-c.TerminateChan:
		err = c.terminate()
		if err == nil {
			err = errors.New(fmt.Sprintf("cmd is terminated, cmd [%s]", c.Cmd))
		}
	}
	return err
}

func (c *LocalCommand) Stderr() string {
	return strings.TrimSpace(string(c.stderr.Bytes()))
}

func (c *LocalCommand) Stdout() string {
	return strings.TrimSpace(string(c.stdout.Bytes()))
}

func (c *LocalCommand) terminate() error {
	if c.Setpgid {
		return syscall.Kill(-c.LocalCommand.Process.Pid, syscall.SIGKILL)
	} else {
		return syscall.Kill(c.LocalCommand.Process.Pid, syscall.SIGKILL)
	}
}
