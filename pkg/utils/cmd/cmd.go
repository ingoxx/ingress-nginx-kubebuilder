package cmd

import (
	"errors"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
)

type Command struct {
	isOutput bool
	args     []string
	name     string
}

func NewCommand(name string, o bool, args []string) Command {
	c := Command{
		name:     name,
		isOutput: o,
		args:     args,
	}

	return c
}

func (c Command) Execute() error {
	var exitError *exec.ExitError
	cmd := exec.Command(c.name, c.args...)
	if c.isOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitError) {
			klog.ErrorS(err, "fail to execute cmd")
			return err
		}
	}

	return nil
}

func (c Command) Output() ([]byte, error) {
	out, err := exec.Command(c.name, c.args...).Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
