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
	// capture the current namespace state and dump it
	// as one YAML file in the respective online (remote)
	// or offline (local) subdirectory:
	namespacestate, err := capture(withstderr, verbose, namespace)
	if err != nil {
		displayerr("Can't capture namespace state", err)
	}
	tsLatest, err = dump(status, namespacestate)
	if err != nil {
		displayerr("Can't dump namespace state", err)
	}
	// only attempt to reconcile if anything has changed:
	if status == prevstatus {
		return tsLatest
	}
	// check which case we have, ONLINE -> OFFLINE or OFFLINE -> ONLINE
	// and respectively switch context (also, make sure remote or local are available):
	switch status {
	case StatusOffline:
		fmt.Printf("Seems I'm %v, will try to switch to local context\n", status)
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, clocal)
		_ = restorefrom(withstderr, verbose, StatusOnline, tsLast)
	case StatusOnline:
		fmt.Printf("Seems I'm %v, switching over to remote context\n", status)
		_ = ensure(status, clocal, cremote)
		_ = use(withstderr, verbose, cremote)
		_ = restorefrom(withstderr, verbose, StatusOffline, tsLast)
	default:
		fmt.Fprintf(os.Stderr, "I don't recognize %v, blame MH9\n", status)
	}
	return tsLatest
}
