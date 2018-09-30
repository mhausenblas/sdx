package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// observeconnection is the connection detector. It tries to do an HTTP GET against
// probeURL and if *anything* comes back we consider ourselves to be online, otherwise
// some network issues prevents us from doing the GET and we are likely offline.
func observeconnection(cremote, clocal string, constat chan string) {
	// the endpoint we're using to check if we're online or offline:
	var probeURL string
	var cctx string
	for {
		cctx = resolvectx(cremote, clocal)
		probeURL = getAPIServerURL(cctx)
		client := http.Client{Timeout: time.Duration(ProbeTimeoutSeconds * time.Second)}
		resp, err := client.Get(probeURL)
		if err != nil {
			fmt.Printf("\x1b[93mConnection detection [%v], probe resulted in:\n%v\x1b[0m\n", StatusOffline, err)
			ccurrent = "local"
			constat <- StatusOffline
			break
		}
		fmt.Printf("\x1b[93mConnection detection [%v], probe %v resulted in:\n%v\x1b[0m\n", StatusOnline, probeURL, resp.Status)
		ccurrent = "remote"
		constat <- StatusOnline
		time.Sleep(CheckConnectionDelaySeconds * time.Second)
	}
}

// getAPIServerURL looks up the API Server url of the kubectx provided.
func getAPIServerURL(kubectx string) string {
	clustername := clusterfromcontext(kubectx)
	probeURL, err := kubectl(false, false, "config", "view",
		"--output=jsonpath='{.clusters[?(@.name == \""+clustername+"\")]..server}'")
	if err != nil {
		displayerr("Can't cuddle the cluster", err)
	}
	probeURL = strings.Trim(probeURL, "'")
	return probeURL
}

// clusterfromcontext extracts the cluster name part from kubectx,
// asssuming it is in the OpenShift format or otherwise assume flat cluster name.
func clusterfromcontext(kubectx string) string {
	switch {
	case strings.Contains(kubectx, "/"):
		// In OpenShift, the context naming format is:
		//  $PROJECT/$CLUSTERNAME/$USER
		// For example:
		//  mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb
		re := regexp.MustCompile("(.*)/(.*)/(.*)")
		return re.FindStringSubmatch(kubectx)[2]
	default:
		return kubectx
	}
}

// resolvectx resolves the context reference, returning
// the actual context to use
func resolvectx(cremote, clocal string) string {
	switch ccurrent {
	case "local":
		return clocal
	case "remote":
		return cremote
	}
	return ""
}
