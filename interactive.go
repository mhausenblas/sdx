package main

import (
	"bufio"
	"fmt"
	"os"
)

// interactivectl enables interactive control, allowing you to use
// commands typed into the terminal to query and/or override the
// current context (manually override).
func interactivectl(namespace, clocal, cremote string, constat chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		switch input {
		case "l", "local", "use local":
			displayfeedback(fmt.Sprintf("Overriding state, switching to local context [%v]", clocal))
			ccurrent = "local"
			constat <- StatusOffline
		case "r", "remote", "use remote":
			displayfeedback(fmt.Sprintf("Overriding state, switching to remote context [%v]", cremote))
			ccurrent = "remote"
			constat <- StatusOnline
		case "s", "status", "show status":
			displayfeedback(fmt.Sprintf("Current status: using %v context, watching namespace %v", ccurrent, namespace))
		default:
		}
	}
}
