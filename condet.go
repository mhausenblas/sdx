package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mhausenblas/kubecuddler"
)

// observeconnection is the connection detector. It performs an HTTP GET against
// probeURL and if *anything* comes back we consider ourselves to be online, otherwise
// some network issues prevent us from doing the GET and we assume we're offline.
func observeconnection(verbose bool, clocal, cremote string, constat chan string) {
	// the endpoint we're using to check if we're online or offline:
	var probeURL string
	for {
		// make sure that when user has manually overridden to use the local
		// context that we disable connection detection with the consequence
		// that until the user selects remote context again no automatic context
		// switch is performed:
		if ccurrent == ContextLocal {
			constat <- StatusOffline
			continue
		}
		probeURL = getAPIServerURL(verbose, cremote)
		client := http.Client{Timeout: time.Duration(ProbeTimeoutSeconds * time.Second)}
		resp, err := client.Get(probeURL)
		if err != nil {
			fmt.Printf("\x1b[93mConnection detection [%v], probe failed\x1b[0m\n", StatusOffline)
			if verbose {
				displayerr("Can't reach remote API Server endpoint", err)
			}
			ccurrent = ContextLocal
			constat <- StatusOffline
			break
		}
		fmt.Printf("\x1b[93mConnection detection [%v], probe %v resulted in: %v\x1b[0m\n", StatusOnline, probeURL, resp.Status)
		ccurrent = ContextRemote
		constat <- StatusOnline
		time.Sleep(CheckConnectionDelaySeconds * time.Second)
	}
}

// getAPIServerURL looks up the API Server URL of the kubectx provided.
func getAPIServerURL(verbose bool, kubectx string) string {
	clustername := clusterfromcontext(kubectx)
	probeURL, err := kubecuddler.Kubectl(false, false, kubectlbin, "config", "view",
		"--output=jsonpath='{.clusters[?(@.name == \""+clustername+"\")]..server}'")
	if err != nil {
		if verbose {
			displayerr("Can't cuddle the cluster", err)
		}
	}
	probeURL = strings.Trim(probeURL, "'")
	return probeURL
}

// clusterfromcontext extracts the cluster name part from the context,
// asssuming it is in the OpenShift format or otherwise assuming a flat cluster name.
func clusterfromcontext(kubectx string) string {
	switch {
	case strings.Contains(kubectx, "/"):
		// In OpenShift, the context naming format is $PROJECT/$CLUSTERNAME/$USER
		// for example: test/api-pro-us-east-1-openshift-com:443/tomjones
		re := regexp.MustCompile("(.*)/(.*)/(.*)")
		return re.FindStringSubmatch(kubectx)[2]
	default:
		return kubectx
	}
}

// resolvectx resolves the context reference, returning the context to use.
func resolvectx(cremote, clocal string) string {
	switch ccurrent {
	case ContextLocal:
		return clocal
	case ContextRemote:
		return cremote
	}
	return ""
}
