package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// capture queries the current state in the active namespace by exporting
// the state of deployments and services as a YAML doc
func capture(withstderr, verbose bool, namespace string) (string, error) {
	yamldoc := "---"
	deploys, err := kubectl(withstderr, verbose, "get", "--namespace="+namespace, "deployments", "--export", "--output=yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't cuddle the cluster due to %v\n", err)
		return "", err
	}
	svcs, err := kubectl(withstderr, verbose, "get", "--namespace="+namespace, "services", "--export", "--output=yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't cuddle the cluster due to %v\n", err)
		return "", err
	}
	yamldoc = deploys + "---\n" + svcs
	return yamldoc, nil
}

// dump stores a YAML doc in a file in:
// $StateCacheDir/$status/
func dump(status, yamldoc string) (string, error) {
	targetdir := filepath.Join(StateCacheDir, status)
	if _, err := os.Stat(targetdir); os.IsNotExist(err) {
		os.Mkdir(targetdir, os.ModePerm)
	}
	ts := time.Now().UnixNano()
	fn := filepath.Join(targetdir, fmt.Sprintf("%v", ts))
	err := ioutil.WriteFile(fn, []byte(yamldoc), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", ts), nil
}

// ensure checks if, depending on the status, the remote or local
// clusters are actually available (in case of local, launches it
//  if this is not the case)
func ensure(status, clocal, cremote string) {
	switch status {
	case StatusOffline:
		fmt.Printf("Attempting to switch to %v, checking if local cluster is available\n", clocal)
		// TODO(mhausenblas): do a "minikube status" or "minishift status" and if not "Running", start it
	case StatusOnline:
		fmt.Printf("Attempting to switch to %v, checking if remote cluster is available \n", cremote)
		// TODO(mhausenblas): do a "kubectl get --raw /api" and if not ready, warn user
	}
}

// restorefrom applies resources from the YAML doc at:
// $StateCacheDir/inv($State)/$TS_LAST
func restorefrom(status, tsLast string) {
	fmt.Printf("Restoring state from %v/%v\n", status, tsLast)
}

// use switches over to provided context as in:
// `kubectl config use-context minikube`
func use(withstderr, verbose bool, context string) error {
	fmt.Printf("Switching over to context %v\n", context)
	_, err := kubectl(withstderr, verbose, "get", "config", "use-context", context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't cuddle the cluster due to %v\n", err)
	}
	return err
}
