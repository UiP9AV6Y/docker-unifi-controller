package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

var version = "dev"

func check(client *http.Client, request *http.Request) error {
	request.Header.Add("User-Agent", "healthcheck/"+version)

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("%s", response.Status)
	}

	return nil
}

func newClient(insecure bool) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport)

	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := &http.Client{
		Transport: transport,
	}

	return client, nil
}

func newRequest(address string, port int, tls bool) (*http.Request, error) {
	socket := strconv.Itoa(port)
	scheme := "http"

	if tls {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s/", scheme, net.JoinHostPort(address, socket))

	return http.NewRequest("GET", url, nil)
}

func main() {
	var client *http.Client
	var request *http.Request
	var port int
	var address string
	var insecure bool
	var verbose bool
	var buildInfo bool
	var tls bool
	var err error

	flag.IntVar(&port, "port", 8080, "webapp listen port")
	flag.StringVar(&address, "address", "127.0.0.1", "webapp listen address")
	flag.BoolVar(&insecure, "insecure", false, "ignore certificate errors")
	flag.BoolVar(&verbose, "verbose", false, "emit additional information")
	flag.BoolVar(&buildInfo, "version", false, "show the application version")
	flag.BoolVar(&tls, "tls", false, "connect to webapp using TLS")

	flag.Parse()

	if buildInfo {
		fmt.Println(version)
		os.Exit(0)
	}

	client, err = newClient(insecure)
	if err != nil {
		fmt.Printf("Invalid client configuration: %v\n", err)
		os.Exit(1)
	}

	request, err = newRequest(address, port, tls)
	if err != nil {
		fmt.Printf("Invalid request configuration: %v\n", err)
		os.Exit(1)
	}

	if err := check(client, request); err != nil {
		fmt.Printf("Webapp (%s:%d) is not reachable: %v\n", address, port, err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Webapp (%s:%d) is accessible\n", address, port)
	}

	os.Exit(0)
}
