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
	// capture the current namespace state and dump it:
	namespacestate, err := capture(withstderr, verbose, namespace, resources)
	if err != nil {
		displayerr("Can't capture namespace state", err)
	}
	displayinfo(fmt.Sprintf("Successfully captured %v in %v", resources, namespace))
	tsLatest, err = dump(status, namespacestate)
	if err != nil {
		displayerr("Can't dump namespace state", err)
	}
	// if nothing changed since previous check, we're done
	if status == prevstatus {
		return tsLast
	}
	// if something changed since previous check, deal with it accordingly:
	cases(status, prevstatus, clocal, cremote, tsLast, withstderr, verbose)
	return tsLatest
}

// cases checks which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
// and respectively switches the context. It also makes sure remote or local are available.
func cases(status, prevstatus, clocal, cremote, tsLast string, withstderr, verbose bool) {
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
