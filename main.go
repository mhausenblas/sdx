package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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

var (
	releaseVersion string
	// what binary to use to speak with the API Server
	// defaults to $(which kubectl)
	kubectlbin string
	// the current context reference, allowed values are `local` and `remote`
	ccurrent string
)

func main() {
	// the namespace I'm trying to keep alive:
	namespace := flag.String("namespace", "default", "the namespace you want me to keep alive")
	// the local context to use:
	clocal := flag.String("local", "minikube", "the local context you want me to use")
	// the remote context to use:
	cremote := flag.String("remote", "", "the remote context you want me to use")
	// defines with which context to start and what to capture:
	policy := flag.String("policy", "local:deployments,services", "defines initial context to use and the kind of resources to capture, there")
	// log level, if verbose is true, give detailed info:
	verbose := flag.Bool("verbose", false, "if set to true, I'll show you all the nitty gritty details")
	// connection status channel, allowed values are StatusXXX
	// this is us used between connection detector and main control loop
	// to communicate the current status:
	constat := make(chan string)
	// the current and previous status, respectively:
	var status, prevstatus string
	// timestamp of most recent dump:
	tsLatest := "0"

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("This is the Kubernetes Seamless Developer Experience (sdx) tool in version %v\n", releaseVersion)
		os.Exit(0)
	}
	// get params and env variables:
	flag.Parse()
	if kb := os.Getenv("SDX_KUBECTL_BIN"); kb != "" {
		kubectlbin = kb
	}
	// make sure we have a remote context to work with:
	if *cremote == "" {
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
	// display config in use:
	showcfg(*clocal, *cremote, *namespace)
	// make sure the initial status is set correctly:
	cinit, resources, err := expandp(*policy)
	if err != nil {
		displayerr("Can't set initial context", err)
		os.Exit(2)
	}
	ccurrent = cinit
	setstate(*clocal, *cremote)
	// launch the connection detector:
	go observeconnection(*verbose, *clocal, *cremote, constat)
	// launch manual override module via keyboard:
	go interactivectl(*namespace, *clocal, *cremote, constat)
	// the main control loop:
	for {
		// read in status from connection detector:
		status = <-constat
		// if prevstatus == "" {
		// 	prevstatus = status
		// }
		// sync state and reconcile, if necessary:
		tsl := syncNReconcile(status, prevstatus, *namespace, *clocal, *cremote, tsLatest, resources, *verbose)
		if tsl != "" {
			tsLatest = tsl
		}
		prevstatus = status
		// wait for next round of sync & reconciliation:
		time.Sleep(SyncStateSeconds * time.Second)
	}
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
