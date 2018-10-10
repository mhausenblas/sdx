package main

import (
	"fmt"
)

// syncNReconcile reconciles the state (local vs. remote).
// If there was a change (local to remote or other way round),
// no matter if triggered by interactive input or connection detection,
// it captures the current state and stores it locally, and then
// switches to the new context.
func syncNReconcile(status, prevstatus, namespace, clocal, cremote, tsLast, resources string, verbose bool) (tsLatest string) {
	withstderr := verbose
	var namespacestate string
	var err error
	if verbose {
		fmt.Printf("Controller sees: current status: [%v] - previous status: [%v] - active context: [%v]\n", status, prevstatus, ccurrent)
	}
	// only act if there was a state change:
	if status == prevstatus {
		return tsLast
	}
	// for valid combinations, capture and store state:
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
	// in any case, switch context and restore state from other context:
	err = switchNRestore(withstderr, verbose, namespace, status, clocal, cremote, tsLast)
	if err != nil {
		if verbose {
			displayerr("No bueno switching context", err)
		}
	}
	return tsLatest
}

// switchNRestore checks which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
// and respectively switches to the context as well as restores the state there.
// Note that it also tries to makes sure remote or local contexts are prepared.
func switchNRestore(withstderr, verbose bool, namespace, status, clocal, cremote, tsLast string) (err error) {
	var res string
	switch status {
	case StatusOffline:
		// TODO(mhausenblas): do a "minikube status" or "minishift status" and if not "Running", start it
		err = use(withstderr, verbose, clocal)
		err = ensure(withstderr, verbose, namespace, status, clocal, cremote)
		res, err = restorefrom(withstderr, verbose, StatusOnline, tsLast)
		if err != nil {
			if verbose {
				displayerr("Can't restore state", err)
			}
		}
		displayinfo(fmt.Sprintf("Successfully switched to [%v] mode and restored resources in %v", status, clocal))
	case StatusOnline:
		// TODO(mhausenblas): do a "kubectl get --raw /api" and if not ready, warn user (?)
		err = use(withstderr, verbose, cremote)
		err = ensure(withstderr, verbose, namespace, status, clocal, cremote)
		res, err = restorefrom(withstderr, verbose, StatusOffline, tsLast)
		if err != nil {
			if verbose {
				displayerr("Can't restore state", err)
			}
		}
		displayinfo(fmt.Sprintf("Successfully switched to [%v] mode and restored resources in %v", status, cremote))
	default:
		if verbose {
			displayerr(fmt.Sprintf("I don't recognize status %v, blame MH9\n", status), nil)
		}
	}
	// show the actual result of the `kubectl apply` call:
	if verbose {
		fmt.Printf("\x1b[34m%v\x1b[0m", res)
	}
	return
}
