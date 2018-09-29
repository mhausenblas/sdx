package main

import (
	"fmt"
	"os"
)

// syncNReconcile syncs the state, reconciles (applies to new environment),
// and switch over to it, IFF there was a change in the status, that is,
// ONLINE -> OFFLINE or other way round.
func syncNReconcile(status, prevstatus, namespace, clocal, cremote, tsLast string, verbose bool) (tsLatest string) {
	withstderr := true
	if status == prevstatus {
		return tsLast
	}
	// check which case we're dealing with and act accordingly:
	cases(status, prevstatus, clocal, cremote, tsLast, withstderr, verbose)
	// capture the current namespace state and dump it:
	namespacestate, err := capture(withstderr, verbose, namespace)
	if err != nil {
		displayerr("Can't capture namespace state", err)
	}
	tsLatest, err = dump(status, namespacestate)
	if err != nil {
		displayerr("Can't dump namespace state", err)
	}
	return tsLatest
}

// cases checks which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
// and respectively switches the context. It also makes sure remote or local are available.
func cases(status, prevstatus, clocal, cremote, tsLast string, withstderr, verbose bool) {
	switch status {
	case StatusOffline:
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, clocal)
		_ = restorefrom(withstderr, verbose, StatusOnline, tsLast)
	case StatusOnline:
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, cremote)
		_ = restorefrom(withstderr, verbose, StatusOffline, tsLast)
	default:
		fmt.Fprintf(os.Stderr, "I don't recognize %v, blame MH9\n", status)
	}
}
