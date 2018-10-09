package main

import (
	"bufio"
	"fmt"
	"os"
)

// manualoverride allows for using keystrokes to override the
// current context, setting the global variable and kubectl it.
func manualoverride(namespace, clocal, cremote string, constat chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		switch input {
		case "l", "local", "use local":
			displayinfo(fmt.Sprintf("Overriding state, switching to local context %v", clocal))
			ccurrent = "local"
			constat <- StatusOffline
		case "r", "remote", "use remote":
			displayinfo(fmt.Sprintf("Overriding state, switching to remote context %v", cremote))
			ccurrent = "remote"
			constat <- StatusOnline
		case "s", "status", "show status":
			displayinfo(fmt.Sprintf("Current status: using %v context, watching namespace %v", ccurrent, namespace))
		default:
		}
	}
}
