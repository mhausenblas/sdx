package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	StatusOnline        = "ONLINE"
	StatusOffline       = "OFFLINE"
	ProbeTimeoutSeconds = 2
)

func main() {
	probeURL := "http://www.google.com"
	fmt.Printf("Starting SDX service\n")
	// the status of the connection
	var constat chan string

	go func() {
		for {
			client := http.Client{Timeout: time.Duration(ProbeTimeoutSeconds * time.Second)}
			resp, err := client.Get(probeURL)
			if err != nil {
				fmt.Printf("Connection detection [%v]\n", StatusOffline)
				constat <- StatusOffline
				break
			}
			fmt.Printf("Connection detection [%v], probe %v resulted in %v \n", StatusOnline, probeURL, resp.Status)
			constat <- StatusOnline
			time.Sleep(2 * time.Second)
		}
	}()
	for {
		msg := <-constat
		fmt.Println(msg)
		time.Sleep(5 * time.Second)
	}
}
