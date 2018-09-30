package main

import (
	"fmt"
	"os"
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

	// capture the current namespace resources depending on the state
	// and if we have a change, try switching over
	if (status == StatusOffline && ccurrent == "local") ||
		(status == StatusOnline && ccurrent == "remote") {
		namespacestate, err = capture(withstderr, verbose, namespace, resources)
		if err != nil {
			displayerr("No bueno capturing state", err)
			return tsLast
		}
		tsLatest, err = dump(status, namespacestate)
		if err != nil {
			displayerr("Can't dump namespace state", err)
		}
		displayinfo(fmt.Sprintf("Successfully backed up %v from namespace %v", resources, namespace))
		// if nothing changed since previous check, we're done
		if status == prevstatus {
			return tsLast
		}
		// if something changed since previous check, deal with it accordingly:
		switchnresurrect(status, clocal, cremote, tsLast, withstderr, verbose)
	}
	if err != nil {
		displayerr(fmt.Sprintf("Can't capture resources from namespace %v", namespace), err)
	}
	return tsLatest
}

// switchnresurrect checks which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
// and respectively switches the context and restores state. It also makes sure remote or local are available.
func switchnresurrect(status, clocal, cremote, tsLast string, withstderr, verbose bool) {
	var res string
	switch status {
	case StatusOffline:
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, clocal)
		res, _ = restorefrom(withstderr, verbose, StatusOnline, tsLast)
	case StatusOnline:
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, cremote)
		res, _ = restorefrom(withstderr, verbose, StatusOffline, tsLast)
	default:
		fmt.Fprintf(os.Stderr, "I don't recognize %v, blame MH9\n", status)
	}
	displayinfo(fmt.Sprintf("Successfully restored resources in %v", status))
	if verbose {
		fmt.Printf("\x1b[34m%v\x1b[0m", res)
	}
}
