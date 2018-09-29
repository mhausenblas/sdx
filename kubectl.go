package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// executes an 'kubectl xxx' command and returns the literal result
func kubectl(withstderr, verbose bool, cmd string, args ...string) (string, error) {
	if kubectlbin == "" {
		bin, err := shellout(withstderr, false, "which", "kubectl")
		if err != nil {
			return "", err
		}
		kubectlbin = bin
	}
	all := append([]string{cmd}, args...)
	result, err := shellout(withstderr, verbose, kubectlbin, all...)
	if err != nil {
		if verbose {
			displayerr("Something went wrong when executing kubectl command", err)
		}
		return "", err
	}
	return result, nil
}

// shells out to execute a command and returns the literal result
func shellout(withstderr, verbose bool, cmd string, args ...string) (string, error) {
	result := ""
	var out bytes.Buffer
	if verbose {
		fmt.Printf("\x1b[94m%v\n", cmd+" "+strings.Join(args, " ")+"\x1b[0m")
	}
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if withstderr {
		c.Stderr = os.Stderr
	}
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		if verbose {
			displayerr("Something went wrong when shelling out", err)
		}
		return "", err
	}
	result = strings.TrimSpace(out.String())
	return result, nil
}
