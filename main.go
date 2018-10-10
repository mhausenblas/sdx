package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// release version provided by the build system, don't touch:
	releaseVersion string
	// what binary to use to speak to API Server, defaults to $(which kubectl):
	kubectlbin string
	// current context reference, allowed values are ContextLocal and ContextRemote:
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
	// connection status channel (allowed values are StatusXXX) used between connection detector
	// and controller to communicate current status:
	constat := make(chan string)
	// the current and previous status, respectively:
	var status, prevstatus string
	// timestamp of most recent dump (only used for orientation purposes and debugging):
	tsLatest := "0"

	// display version when ask and exit:
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
		handlenoremote()
	}
	// make sure local cache is ready:
	initcache(*verbose)
	// make sure we don't leave some crap state in the local cache
	// when user exits via CTRL+C:
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		emptycache(*verbose)
		os.Exit(0)
	}()
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
	// launch interactive control (via keyboard):
	go interactivectl(*namespace, *clocal, *cremote, constat)
	// the main control loop:
	for {
		// receive status from connection detector:
		status = <-constat
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
