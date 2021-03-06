package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mhausenblas/kubecuddler"
)

// initcache sets up $StateCacheDir/$state/ directories.
func initcache(verbose bool) {
	targetdir := filepath.Join(StateCacheDir, StatusOffline)
	if _, err := os.Stat(targetdir); os.IsNotExist(err) {
		e := os.MkdirAll(targetdir, os.ModePerm)
		if err != nil {
			if verbose {
				displayerr("Can't create state cache", e)
			}
		}
	}
	targetdir = filepath.Join(StateCacheDir, StatusOnline)
	if _, err := os.Stat(targetdir); os.IsNotExist(err) {
		e := os.MkdirAll(targetdir, os.ModePerm)
		if err != nil {
			if verbose {
				displayerr("Can't create state cache", e)
			}
		}
	}
}

// ensure checks if, depending on the status, the remote or local
// context are set up correctly, for example, if the namespace exists locally.
func ensure(withstderr, verbose bool, namespace, status, clocal, cremote string) error {
	switch status {
	case StatusOffline:
		if verbose {
			displayinfo(fmt.Sprintf("Checking if local context [%v] is ready", clocal))
		}
		// make sure that if the namespace exists we re-create it:
		_, err := kubecuddler.Kubectl(withstderr, verbose, kubectlbin, "get", "namespace", namespace)
		if err == nil {
			_, err := kubecuddler.Kubectl(false, verbose, kubectlbin, "delete", "namespace", namespace)
			if err != nil {
				return err
			}
		}
		_, err = kubecuddler.Kubectl(false, verbose, kubectlbin, "create", "namespace", namespace)
		if err != nil {
			return err
		}
		displayinfo(fmt.Sprintf("Recreated namespace [%v] in local context", namespace))
	case StatusOnline:
		if verbose {
			displayinfo(fmt.Sprintf("Checking if remote context [%v] is ready", cremote))
		}
	}
	return nil
}

// capture queries the current state in the active namespace by exporting
// the state of deployments and services as a YAML doc
func capture(withstderr, verbose bool, namespace, resources string) (string, error) {
	yamldoc := "# This file has been automatically generated by kube-sdx\n\n---\n"
	if !strings.Contains(resources, ",") {
		return "", fmt.Errorf("Invalid resource list, must be of format: '$RESOURCE_KIND1,$RESOURCE_KIND2,...'")
	}
	resourcelist := strings.Split(resources, ",")
	for _, reskind := range resourcelist {
		reskind = strings.TrimSpace(reskind)
		yamlfrag, err := kubecuddler.Kubectl(withstderr, verbose, kubectlbin, "get", "--namespace="+namespace, reskind, "--export", "--output=yaml")
		if err != nil {
			return "", err
		}
		yamldoc += yamlfrag + "\n---\n"
	}
	return yamldoc, nil
}

// dump stores a YAML doc in a file at $StateCacheDir/$status/$StateFile
// It returns the timestamp  in Unix time of when the file was written.
func dump(status, yamldoc string) (string, error) {
	targetdir := filepath.Join(StateCacheDir, status)
	ts := time.Now().UnixNano()
	fn := filepath.Join(targetdir, StateFile)
	// make sure we drop the cluster IP spec field for services:
	re := regexp.MustCompile("(?m)[\r\n]+^.*clusterIP:.*$")
	yamldoc = re.ReplaceAllString(yamldoc, "")
	// write out resulting YAML doc to file:
	err := ioutil.WriteFile(fn, []byte(yamldoc), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", ts), nil
}

// restorefrom applies resources from the YAML doc at $StateCacheDir/$state/$StateFile
func restorefrom(withstderr, verbose bool, state, tsLast string) (res string, err error) {
	statefile := filepath.Join(StateCacheDir, state, StateFile)
	if verbose {
		displayinfo(fmt.Sprintf("Trying to restore state from %v/%v@%v", state, StateFile, tsLast))
	}
	if _, err = os.Stat(statefile); !os.IsNotExist(err) {
		res, err = kubecuddler.Kubectl(withstderr, verbose, kubectlbin, "apply", "--filename="+statefile)
		if err != nil {
			return "", err
		}
	}
	return res, err
}

// use switches over to the provided context using `kubectl config use-context`.
func use(withstderr, verbose bool, context string) error {
	_, err := kubecuddler.Kubectl(withstderr, verbose, kubectlbin, "config", "use-context", context)
	if err != nil {
		return err
	}
	displayinfo(fmt.Sprintf("Now using context [%v]", context))
	return nil
}

// emptycache deletes the state files in $StateCacheDir/$state/*
func emptycache(verbose bool) {
	targetdir := filepath.Join(StateCacheDir, StatusOffline)
	err := os.Remove(filepath.Join(targetdir, StateFile))
	if err != nil {
		if verbose {
			displayerr("\nCan't clean up state", err)
		}
	}
	targetdir = filepath.Join(StateCacheDir, StatusOnline)
	err = os.Remove(filepath.Join(targetdir, StateFile))
	if err != nil {
		if verbose {
			displayerr("\nCan't clean up state", err)
		}
	}
	displayinfo("\nNuked local cache, all state gone. Thanks for using kube-sdx and have a nice day! :)")
}
