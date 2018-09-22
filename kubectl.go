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
	bin, err := shellout(withstderr, false, "which", "kubectl")
	if err != nil {
		return "", err
	}
	all := append([]string{cmd}, args...)
	result, err := shellout(withstderr, verbose, bin, all...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return "", err
	}
	return result, nil
}

// shells out to execute a command and returns the literal result
func shellout(withstderr, verbose bool, cmd string, args ...string) (string, error) {
	result := ""
	var out bytes.Buffer
	if verbose {
		fmt.Printf("%v\n", cmd+" "+strings.Join(args, " "))
	}
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if withstderr {
		c.Stderr = os.Stderr
	}
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return "", err
	}
	result = strings.TrimSpace(out.String())
	return result, nil
}
