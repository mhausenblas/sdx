package main

import (
	"fmt"
	"net/http"
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
)

func main() {
	// the endpoint we're using to check if we're online or offline
	// TODO(mhausenblas): change to API server address or make it configurable?
	probeURL := "http://www.google.com"
	// the status of the connection, can be StatusXXX
	constat := make(chan string)
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
		msg := <-constat
		syncNReconcile(msg)
		// wait for next round of sync & reconciliation:
		time.Sleep(SyncStateSeconds * time.Second)
	}
}

func syncNReconcile(status string) {
	switch status {
	case StatusOffline:
		fmt.Printf("Seems I'm %v, will try to switch over to local env\n", status)
	case StatusOnline:
		fmt.Printf("Seems I'm %v, will sync state and switch over to remote env\n", status)
	default:
		fmt.Printf("I don't recognize %v, blame MH9\n", status)
	}
}
