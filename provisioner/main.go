package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
)

const (
	defaultImage  = "a3da5530-71c6-4405-b64f-fd2da99d303c" // Ubuntu 12.04 LTS (Precise Pangolin) (PVHVM)
	defaultFlavor = "general1-1"                           // 1 GB General Purpose v1
	cachePath     = "/tmp/sandbox_auth_token"
)

type clientList map[string]*gophercloud.ServiceClient

func setupClients() clientList {
	// Check if cache exists
	if _, err := os.Stat(cachePath); err == nil {
		jsonBytes, err := ioutil.ReadFile(cachePath)
		if string(jsonBytes) != "" && err == nil {
			var cache struct {
				TokenID string `json:"token"`
				LB      string `json:"lbEndpoint"`
				CM      string `json:"cmEndpoint"`
			}
			err := json.Unmarshal(jsonBytes, &cache)
			checkErr("unmarshalling JSON cache", err)

			if cache.TokenID != "" && cache.LB != "" && cache.CM != "" {
				provider := &gophercloud.ProviderClient{
					IdentityBase: os.Getenv("RS_AUTH_URL"),
					TokenID:      cache.TokenID,
				}
				return clientList{
					"compute": &gophercloud.ServiceClient{ProviderClient: provider, Endpoint: cache.CM},
					"lb":      &gophercloud.ServiceClient{ProviderClient: provider, Endpoint: cache.LB},
				}
			}
		}
	}

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

	jsonStr, err := json.Marshal(map[string]string{
		"token":      client.TokenID,
		"lbEndpoint": lb.Endpoint,
		"cmEndpoint": compute.Endpoint,
	})
	checkErr("marshalling JSON cache", err)

	ioutil.WriteFile(cachePath, jsonStr, 0644)

	return clientList{"compute": compute, "lb": lb}
}

func main() {
	http.HandleFunc("/images", handleImages)
	http.HandleFunc("/flavors", handleFlavors)
	http.HandleFunc("/create_server", handleServerCreate)
	http.HandleFunc("/create_lb", handleLBCreate)

	http.ListenAndServe(":8080", nil)
}
