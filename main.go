package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
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

func main() {
	// the namespace I'm trying to keep alive:
	namespace := flag.String("namespace", "default", "the namespace you want me to keep alive")
	// the local context to use
	clocal := flag.String("local", "minikube", "the local context you want me to use")
	// the remote context to use
	cremote := flag.String("remote", "", "the remote context you want me to use")
	// the endpoint we're using to check if we're online or offline
	// which is by default the API server address from the remote context
	var probeURL string
	// connection status channel, allowed values are StatusXXX
	// this is us used between connection detector and main control loop
	// to communicate the current status
	constat := make(chan string)
	// the current and previous status, respectively
	var status, prevstatus string
	// timestamp of most recent dump
	var tsLatest string

	flag.Parse()

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
	showcfg(*clocal, *cremote, *namespace)
	// the connection detector, simply tries to do an HTTP GET against probeURL
	// and if *anything* comes back we consider ourselves to be online, otherwise
	// some network issues prevents us from doing the GET and we are likely offline.
	go func() {
		for {
			clustername := clusterfromcontext(*cremote)
			probeURL, err = kubectl(false, false, "config", "view",
				"--output=jsonpath='{.clusters[?(@.name == \""+clustername+"\")]..server}'")
			if err != nil {
				displayerr("Can't cuddle the cluster", err)
				os.Exit(1)
			}
			probeURL = strings.Trim(probeURL, "'")
			// fmt.Printf("Trying to probe URL %v\n", probeURL)
			client := http.Client{Timeout: time.Duration(ProbeTimeoutSeconds * time.Second)}
			resp, err := client.Get(probeURL)
			if err != nil {
				fmt.Printf("Connection detection [%v], probe resulted in %v\n", StatusOffline, err)
				constat <- StatusOffline
				continue
			}
			fmt.Printf("Connection detection [%v], probe %v resulted in %v \n", StatusOnline, probeURL, resp.Status)
			constat <- StatusOnline
			time.Sleep(CheckConnectionDelaySeconds * time.Second)
		}
	}()
	// the main control loop
	for {
		// read in status from connection detector:
		status = <-constat
		if prevstatus == "" {
			prevstatus = status
		}
		// sync state and reconcile, if necessary:
		tsl := syncNReconcile(status, prevstatus, *namespace, *clocal, *cremote, tsLatest)
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
