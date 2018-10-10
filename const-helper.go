package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mhausenblas/kubecuddler"
)

const (
	// StatusOnline signals we have access to the remote cluster:
	StatusOnline = "ONLINE"
	// StatusOffline signals we do not have access to the remote cluster:
	StatusOffline = "OFFLINE"
	// ContextLocal represents the local context
	ContextLocal = "local"
	// ContextRemote represents the remote context
	ContextRemote = "remote"
	// ProbeTimeoutSeconds defines how long to try to get a result when probing:
	ProbeTimeoutSeconds = 5
	// CheckConnectionDelaySeconds defines how long to wait between two connection checks:
	CheckConnectionDelaySeconds = 2
	// SyncStateSeconds defines how long to wait between two reconcile rounds:
	SyncStateSeconds = 2
	// StateCacheDir defines the directory to use to keep track of the local and remote state:
	StateCacheDir = "/tmp/kube-sdx"
	// StateFile defines the name of the file to keep track of the local and remote state:
	StateFile = "latest.yaml"
)

// handlenoremote warns user no remote is given and provides suggestions to use.
func handlenoremote() {
	fmt.Println("\x1b[91mI'm sorry Dave, I'm afraid I can't do that.\x1b[0m")
	fmt.Println("I need to know which remote context you want, pick one from below and provide it via the \x1b[1m--remote\x1b[0m parameter:\n")
	contexts, err := kubecuddler.Kubectl(false, false, kubectlbin, "config", "get-contexts")
	if err != nil {
		displayerr("Can't cuddle the Kube", err)
		os.Exit(1)
	}
	fmt.Println(contexts)
	os.Exit(2)
}

// setstate sets the context directly
func setstate(clocal, cremote string) {
	newcontext := ""
	switch ccurrent {
	case ContextLocal:
		newcontext = clocal
	case ContextRemote:
		newcontext = cremote
	default:
		displayerr("I don't know about the context reference", fmt.Errorf(ccurrent))
		os.Exit(2)
	}
	err := use(false, false, newcontext)
	if err != nil {
		displayerr("Can't cuddle the Kube", err)
		os.Exit(2)
	}
}

// showcfg prints the current config to screen
func showcfg(clocal, cremote, namespace string) {
	fmt.Println("\n*** STARTING SDX\n")
	fmt.Printf("I'm using the following configuration:\n")
	fmt.Printf("- local context: \x1b[96m%v\x1b[0m\n", clocal)
	fmt.Printf("- remote context: \x1b[96m%v\x1b[0m\n", cremote)
	fmt.Printf("- namespace to keep alive: \x1b[96m%v\x1b[0m\n", namespace)
	fmt.Println("***\n")
}

// expandp checks if we're dealing with a valid policy string and if so,
// extracts the initial context and the (comma-separated list of) resources
// that we're supposed to capture, back up and restore again.
func expandp(policy string) (cinit, resources string, err error) {
	if !strings.Contains(policy, ":") {
		return "", "", fmt.Errorf("Invalid policy, must be of format: '$CONTEXT:$RESOURCE*'")
	}
	cinit, resources = strings.Split(policy, ":")[0], strings.Split(policy, ":")[1]
	return cinit, resources, nil
}

// displayinfo writes message to stdout
func displayinfo(msg string) {
	_, _ = fmt.Fprintf(os.Stdout, "\x1b[32m%v\x1b[0m\n", msg)
}

// displayfeedback writes message to stdout
func displayfeedback(msg string) {
	_, _ = fmt.Fprintf(os.Stdout, "\x1b[35m%v\x1b[0m\n", msg)
}

// displayerr writes message and error out to stderr
func displayerr(msg string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%v: \x1b[91m%v\x1b[0m\n", msg, err)
}
