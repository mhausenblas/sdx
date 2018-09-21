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
	// the status of the connection, can be StatusXXX
	constat := make(chan string)

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
	for {
		// read in status from connection detector
		status := <-constat
		syncNReconcile(status, *namespace)
		// wait for next round of sync & reconciliation:
		time.Sleep(SyncStateSeconds * time.Second)
	}
}

func syncNReconcile(status, namespace string) {
	switch status {
	case StatusOffline:
		fmt.Printf("Seems I'm %v, will try to switch over to local env\n", status)
	case StatusOnline:
		fmt.Printf("Seems I'm %v, will sync state and switch over to remote env\n", status)
		r, err := kubectl(true, "get", "--namespace="+namespace, "deployments", "--export", "--output=yaml")
		if err != nil {
			fmt.Printf("Can't cuddle the cluster due to %v\n", err)
			return
		}
		err = dump(r, "deployments")
		if err != nil {
			fmt.Printf("Can't dump state due to %v\n", err)
			return
		}
	default:
		fmt.Printf("I don't recognize %v, blame MH9\n", status)
	}
}

func dump(yamlblob, context string) error {
	if _, err := os.Stat(StateCacheDir); os.IsNotExist(err) {
		os.Mkdir(StateCacheDir, os.ModePerm)
	}
	ts := time.Now().UnixNano()
	fn := filepath.Join(StateCacheDir, fmt.Sprintf("%v_%v", ts, context))
	err := ioutil.WriteFile(fn, []byte(yamlblob), 0644)
	if err != nil {
		return err
	}
	return nil
}
