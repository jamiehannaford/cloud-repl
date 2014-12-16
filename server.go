package main

import (
	"fmt"
	"net/http"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

func createServer(client *gophercloud.ServiceClient, content *string, count int) string {
	name := randomStr("sandbox_", 10)

	opts := &servers.CreateOpts{
		Name:       name,
		ImageRef:   defaultImage,
		FlavorRef:  defaultFlavor,
		DiskConfig: "MANUAL",
	}

	server, err := servers.Create(client, opts).Extract()
	checkErr("creating server", err)

	err = servers.WaitForStatus(client, server.ID, "ACTIVE", 60)
	checkErr("waiting for server to boot", err)

	if count == 1 {
		*content += fmt.Sprintf("| %-20s | %-36s | %-15s |\n", "Name", "ID", "Password")
		*content += fmt.Sprintf("|%s|%s|%s|\n", hyphens(20), hyphens(36), hyphens(15))
	}

	*content += fmt.Sprintf("| %-20s | %-36s | %-15s | \n", name, server.ID, server.AdminPass)

	return server.ID
}

func handleServerCreate(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")

	client := setupClients()["compute"]
	content := ""

	for i := 1; i <= 2; i++ {
		id := createServer(client, &content, i)
		w.Header().Set("New-Servers", id)
	}

	fmt.Fprintf(w, content)
	w.Header().Set("Blah", "*")
}
