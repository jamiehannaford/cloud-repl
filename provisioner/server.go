package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

func createServer(client *gophercloud.ServiceClient, content *string, name, prefix string) string {
	opts := &servers.CreateOpts{
		Name:       prefix + name,
		ImageRef:   defaultImage,
		FlavorRef:  defaultFlavor,
		DiskConfig: "MANUAL",
	}

	server, err := servers.Create(client, opts).Extract()
	checkErr("creating server", err)

	err = servers.WaitForStatus(client, server.ID, "ACTIVE", 60)
	checkErr("waiting for server to boot", err)

	*content += fmt.Sprintf("| %-20s | %-36s | %-15s |\n", "Name", "ID", "Password")
	*content += fmt.Sprintf("|%s|%s|%s|\n", hyphens(20), hyphens(36), hyphens(15))
	*content += fmt.Sprintf("| %-20s | %-36s | %-15s | \n", name, server.ID, server.AdminPass)

	return server.ID
}

func getPrefix(r *http.Request) string {
	prefix := r.Header.Get("User-Prefix")
	if prefix == "" {
		prefix = randomStr("sandbox_", 10)
	}
	return prefix + "_"
}

func getName(r *http.Request) string {
	decoder := json.NewDecoder(r.Body)

	body := struct {
		Name string `json:"name"`
	}{}

	err := decoder.Decode(&body)
	checkErr("decoding create_server request JSON", err)

	if body.Name == "" {
		return randomStr("sandbox_", 10)
	}

	return body.Name
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

	ip := createServer(client, &content, getName(r), getPrefix(r))
	w.Header().Set("Server-IP", ip)

	fmt.Fprintf(w, content)
}