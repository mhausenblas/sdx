package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	// the endpoint we're using to check if we're online or offline
	// TODO(mhausenblas): change to API server address or make it configurable?
	probeURL := "http://www.google.com"
	// connection status channel, allowed values are StatusXXX
	constat := make(chan string)
	// the current and previous status
	var status, prevstatus string

	flag.Parse()

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
		syncNReconcile(status, prevstatus, *namespace)
		prevstatus = status
		// wait for next round of sync & reconciliation:
		time.Sleep(SyncStateSeconds * time.Second)
	}
}

// syncNReconcile syncs the state, reconciles (applies to new environment),
// and switch over to it, IFF there was a change in the status, that is,
// ONLINE -> OFFLINE or other way round.
func syncNReconcile(status, prevstatus, namespace string) {
	// only attempt to sync and reconcile if anything has changed:
	if status == prevstatus {
		return
	}
	// check which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
	switch status {
	case StatusOffline:
		fmt.Printf("Seems I'm %v, will try to switch over to local env\n", status)
		ensurelocal()
		restore()
		selectcontext()
	case StatusOnline:
		fmt.Printf("Seems I'm %v, will sync state and switch over to remote env\n", status)
		r, err := kubectl(true, "get", "--namespace="+namespace, "deployments", "--export", "--output=yaml")
		if err != nil {
			fmt.Printf("Can't cuddle the cluster due to %v\n", err)
			return
		}
		err = dump("deployments", r)
		if err != nil {
			fmt.Printf("Can't dump state due to %v\n", err)
			return
		}
	default:
		fmt.Printf("I don't recognize %v, blame MH9\n", status)
	}
}

// stores a YAML doc in a file in format timestamp + resource kind
func dump(reskind, yamlblob string) error {
	if _, err := os.Stat(StateCacheDir); os.IsNotExist(err) {
		os.Mkdir(StateCacheDir, os.ModePerm)
	}
	ts := time.Now().UnixNano()
	fn := filepath.Join(StateCacheDir, fmt.Sprintf("%v_%v", ts, context))
	err := ioutil.WriteFile(fn, []byte(yamlblob), 0644)
	return err
}

// checks if Minikube or Minishift is running and if not, launches it
func ensurelocal() {

}

// applies resources from $StateCacheDir/inv($State)/$TS_$RESKIND
func restore() {

}

// switches over to local context, like `kubectl config use-context minikube`
func selectcontext() {

}
