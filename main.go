package main

import (
	"net/http"
	"os"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
)

const (
	defaultImage  = "a3da5530-71c6-4405-b64f-fd2da99d303c" // Ubuntu 12.04 LTS (Precise Pangolin) (PVHVM)
	defaultFlavor = "general1-1"                           // 1 GB General Purpose v1
)

type clientList map[string]*gophercloud.ServiceClient

func setupClients() clientList {
	opts, err := rackspace.AuthOptionsFromEnv()
	checkErr("retrieving env vars", err)

	region := os.Getenv("RS_REGION")

	client, err := rackspace.AuthenticatedClient(opts)
	checkErr("authenticating", err)

	eopts := gophercloud.EndpointOpts{Region: region}

	compute, err := rackspace.NewComputeV2(client, eopts)
	checkErr("creating compute service", err)

	lb, err := rackspace.NewLBV1(client, eopts)
	checkErr("creating LB service", err)

	return clientList{"compute": compute, "lb": lb}
}

func main() {
	http.HandleFunc("/images", handleImages)
	http.HandleFunc("/flavors", handleFlavors)
	http.HandleFunc("/create_server", handleServerCreate)
	http.HandleFunc("/create_lb", handleLBCreate)

	http.ListenAndServe(":8080", nil)
}
