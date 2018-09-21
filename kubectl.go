package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// executes an 'kubectl xxx' command and returns the literal result
func kubectl(withstderr bool, cmd string, args ...string) (string, error) {
	bin, err := shellout(withstderr, "which", "kubectl")
	if err != nil {
		return "", err
	}
	all := append([]string{cmd}, args...)
	result, err := shellout(withstderr, bin, all...)
	if err != nil {
		return "", err
	}
	return result, nil
}

// shells out to execute a command and returns the literal result
func shellout(withstderr bool, cmd string, args ...string) (string, error) {
	result := ""
	var out bytes.Buffer
	fmt.Printf("%v\n", cmd+" "+strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if withstderr {
		c.Stderr = os.Stderr
	}
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		return result, err
	}
	result = strings.TrimSpace(out.String())
	return result, nil
}
