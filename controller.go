package main

import (
	"fmt"
)

// syncNReconcile syncs the state, reconciles (applies to new environment),
// and switch over to it, IFF there was a change in the status, that is,
// ONLINE -> OFFLINE or other way round.
func syncNReconcile(status, prevstatus, namespace, clocal, cremote, tsLast, resources string, verbose bool) (tsLatest string) {
	withstderr := verbose
	var namespacestate string
	var err error
	if verbose {
		fmt.Printf("Controller sees: \x1b[92mstatus: %v prev status: %v context: %v\x1b[0m\n", status, prevstatus, ccurrent)
	}

	if status == prevstatus {
		return tsLast
	}

	if (status == StatusOffline && ccurrent == "local") ||
		(status == StatusOnline && ccurrent == "remote") {
		namespacestate, err = capture(withstderr, verbose, namespace, resources)
		if err != nil {
			if verbose {
				displayerr("No bueno capturing state", err)
			}
		}
		tsLatest, err = dump(status, namespacestate)
		if err != nil {
			if verbose {
				displayerr("Can't dump namespace state", err)
			}
		}
		if tsLatest != "" {
			displayinfo(fmt.Sprintf("Successfully backed up %v from namespace %v", resources, namespace))
		}
	}

	err = switchnresurrect(withstderr, verbose, namespace, status, clocal, cremote, tsLast)
	if err != nil {
		if verbose {
			displayerr("No bueno switching context", err)
		}
	}
	return tsLatest
}

// switchnresurrect checks which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
// and respectively switches the context and restores state. It also makes sure remote or local are available.
func switchnresurrect(withstderr, verbose bool, namespace, status, clocal, cremote, tsLast string) (err error) {
	var res string
	switch status {
	case StatusOffline:
		// TODO(mhausenblas): do a "minikube status" or "minishift status" and if not "Running", start it
		err = use(withstderr, verbose, clocal)
		err = ensure(withstderr, verbose, namespace, status, clocal, cremote)
		res, err = restorefrom(withstderr, verbose, StatusOnline, tsLast)
	case StatusOnline:
		// TODO(mhausenblas): do a "kubectl get --raw /api" and if not ready, warn user
		err = use(withstderr, verbose, cremote)
		err = ensure(withstderr, verbose, namespace, status, clocal, cremote)
		res, err = restorefrom(withstderr, verbose, StatusOffline, tsLast)
	default:
		if verbose {
			displayerr(fmt.Sprintf("I don't recognize %v, blame MH9\n", status), nil)
		}
	}
	displayinfo(fmt.Sprintf("Successfully restored resources in %v", status))
	if verbose {
		fmt.Printf("\x1b[34m%v\x1b[0m", res)
	}
	return
}
