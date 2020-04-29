package cmd

import (
	"bytes"
	"time"
)

type BaseCommand struct {
	Cmd           string
	Timeout       time.Duration
	TerminateChan chan int
	stdout        bytes.Buffer
	stderr        bytes.Buffer
	state         string
}
