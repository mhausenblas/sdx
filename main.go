package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	// representing the online state:
	StatusOnline = "ONLINE"
	// representing the offline state:
	StatusOffline = "OFFLINE"
	// how long to try to get a result when probing:
	ProbeTimeoutSeconds = 10
	// how quick to check connection:
	CheckConnectionDelaySeconds = 2
	// how quick to sync reconcile state:
	SyncStateSeconds = 2
	// the directory to keep track of the state:
	StateCacheDir = "/tmp/kube-sdx"
)

func main() {
	// the namespace we're operating on (sync & reconcile)
	namespace := flag.String("namespace", "default", "the namespace you want to keep alive")
	// the local context to use
	clocal := flag.String("local", "minikube", "the local context to use")
	// the remote context to use
	cremote := flag.String("remote", "", "the remote context to use")
	// the endpoint we're using to check if we're online or offline
	// TODO(mhausenblas): change to API server address or make it configurable?
	probeURL := "http://www.google.com"
	// connection status channel, allowed values are StatusXXX
	constat := make(chan string)
	// the current and previous status
	var status, prevstatus string
	// timestamp of most recent dump
	var tsLatest string

	flag.Parse()

	// make sure we have a remote context to work with:
	if *cremote == "" {
		fmt.Printf("\x1b[91mI'm sorry Dave, I'm afraid I can't do that.\n\x1b[0m")
		fmt.Printf("I need to know which remote context you want, pick one from below and provide it via the \x1b[1m--remote\x1b[0m parameter:\n\n")
		contexts, _ := kubectl(false, false, "config", "get-contexts")
		fmt.Println(contexts)
		os.Exit(1)
	}
	showcfg(*clocal, *cremote, *namespace)
	// the connection detector, simply tries to do an HTTP GET against probeURL
	// and if *anything* comes back we consider ourselves to be online, otherwise
	// some network issues prevents us from doing the GET and we are likely offline.
	go func() {
		for {
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
	// the main control loop:
	for {
		// read in status from connection detector:
		status = <-constat
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

// syncNReconcile syncs the state, reconciles (applies to new environment),
// and switch over to it, IFF there was a change in the status, that is,
// ONLINE -> OFFLINE or other way round.
func syncNReconcile(status, prevstatus, namespace, clocal, cremote, tsLast string) (tsLatest string) {
	withstderr := true
	verbose := false
	// only attempt to sync and reconcile if anything has changed:
	if status == prevstatus {
		return ""
	}
	// capture the current namespace state and dump it
	// as one YAML file in the respective online (remote)
	// or offline (local) subdirectory:
	namespacestate, err := capture(withstderr, verbose, namespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't capture namespace state due to %v\n", err)
		return ""
	}
	tsLatest, err = dump(status, namespacestate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't dump namespace state due to %v\n", err)
		return ""
	}
	// check which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
	// and respectively switch context (also, make sure remote or local are available):
	switch status {
	case StatusOffline:
		fmt.Printf("Seems I'm %v, will try to switch to local context\n", status)
		ensure(status, clocal, cremote)
		restorefrom(StatusOnline, tsLast)
		use(clocal)
	case StatusOnline:
		fmt.Printf("Seems I'm %v, switching over to remote context\n", status)
		ensure(status, clocal, cremote)
		restorefrom(StatusOffline, tsLast)
		use(cremote)
	default:
		fmt.Fprintf(os.Stderr, "I don't recognize %v, blame MH9\n", status)
	}
	return
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
