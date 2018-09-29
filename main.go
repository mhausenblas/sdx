package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"
)

const (
	// StatusOnline signals we have access to the remote cluster:
	StatusOnline = "ONLINE"
	// StatusOffline signals we do not have access to the remote cluster:
	StatusOffline = "OFFLINE"
	// ProbeTimeoutSeconds defines how long to try to get a result when probing:
	ProbeTimeoutSeconds = 10
	// CheckConnectionDelaySeconds defines how long to wait between two connection checks:
	CheckConnectionDelaySeconds = 2
	// SyncStateSeconds defines how long to wait between two reconcile rounds:
	SyncStateSeconds = 2
	// StateCacheDir defines the directory to use to keep track of the local and remote state:
	StateCacheDir = "/tmp/kube-sdx"
)

var (
	// what binary to use to speak with the API Server
	// defaults to $(which kubectl)
	kubectlbin string
)

func main() {
	// the namespace I'm trying to keep alive:
	namespace := flag.String("namespace", "default", "the namespace you want me to keep alive")
	// the local context to use
	clocal := flag.String("local", "minikube", "the local context you want me to use")
	// the remote context to use
	cremote := flag.String("remote", "", "the remote context you want me to use")
	// log level
	verbose := flag.Bool("verbose", false, "if set to true, I'll show you all the nitty gritty details")
	// connection status channel, allowed values are StatusXXX
	// this is us used between connection detector and main control loop
	// to communicate the current status
	constat := make(chan string)
	// the current and previous status, respectively
	var status, prevstatus string
	// timestamp of most recent dump
	var tsLatest string

	flag.Parse()

	if kb := os.Getenv("SDX_KUBECTL_BIN"); kb != "" {
		kubectlbin = kb
	}

	// make sure we have a remote context to work with:
	if *cremote == "" {
		fmt.Println("\x1b[91mI'm sorry Dave, I'm afraid I can't do that.\x1b[0m")
		fmt.Println("I need to know which remote context you want, pick one from below and provide it via the \x1b[1m--remote\x1b[0m parameter:\n")
		contexts, err := kubectl(false, false, "config", "get-contexts")
		if err != nil {
			displayerr("Can't cuddle the cluster", err)
			os.Exit(1)
		}
		fmt.Println(contexts)
		os.Exit(2)
	}
	err := use(false, false, *cremote)
	if err != nil {
		displayerr("Can't cuddle the cluster", err)
		os.Exit(1)
	}
	// display config in use:
	showcfg(*clocal, *cremote, *namespace)
	// the connection detector:
	go observeconnection(*cremote, constat)
	// the main control loop:
	for {
		// read in status from connection detector:
		status = <-constat
		if prevstatus == "" {
			prevstatus = status
		}
		// sync state and reconcile, if necessary:
		tsl := syncNReconcile(status, prevstatus, *namespace, *clocal, *cremote, tsLatest, *verbose)
		if tsl != "" {
			tsLatest = tsl
		}
		prevstatus = status
		// wait for next round of sync & reconciliation:
		time.Sleep(SyncStateSeconds * time.Second)
	}
}

// showcfg prints the current config to screen
func showcfg(clocal, cremote, namespace string) {
	fmt.Println("--- STARTING SDX\n")
	fmt.Printf("I'm using the following configuration:\n")
	fmt.Printf("- local context: \x1b[34m%v\x1b[0m\n", clocal)
	fmt.Printf("- remote context: \x1b[34m%v\x1b[0m\n", cremote)
	fmt.Printf("- namespace to keep alive: \x1b[34m%v\x1b[0m\n", namespace)
	fmt.Println("---\n")
}

// clusterfromcontext extracts the cluster name part from
// a context name, asssuming it is in the OpenShift format.
func clusterfromcontext(context string) string {
	// In OpenShift, the context naming format is:
	// $PROJECT/$CLUSTERNAME/$USER for example:
	// mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb
	re := regexp.MustCompile("(.*)/(.*)/(.*)")
	return re.FindStringSubmatch(context)[2]
}

// displayerr write message and error out to stderr
func displayerr(msg string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%v: \x1b[91m%v\x1b[0m\n", msg, err)
}
