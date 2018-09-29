package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// observeconnection is the connection detector. It tries to do an HTTP GET against
// probeURL and if *anything* comes back we consider ourselves to be online, otherwise
// some network issues prevents us from doing the GET and we are likely offline.
func observeconnection(cremote string, constat chan string) {
	// the endpoint we're using to check if we're online or offline
	// which is by default the API server address from the remote context
	var probeURL string
	var err error
	for {
		clustername := clusterfromcontext(cremote)
		probeURL, err = kubectl(false, false, "config", "view",
			"--output=jsonpath='{.clusters[?(@.name == \""+clustername+"\")]..server}'")
		if err != nil {
			displayerr("Can't cuddle the cluster", err)
			os.Exit(1)
		}
		probeURL = strings.Trim(probeURL, "'")
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
}
