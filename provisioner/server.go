package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

func createServer(client *gophercloud.ServiceClient, content *string, name string) string {
	opts := &servers.CreateOpts{
		Name:       randomStr("sandbox_", 10) + name,
		ImageRef:   defaultImage,
		FlavorRef:  defaultFlavor,
		DiskConfig: "MANUAL",
	}

	server, err := servers.Create(client, opts).Extract()
	checkErr("creating server", err)

	ip := ""

	err = gophercloud.WaitFor(60, func() (bool, error) {
		current, err := servers.Get(client, server.ID).Extract()
		ip = current.AccessIPv4
		if err != nil {
			return false, err
		}
		if current.Status == "ACTIVE" {
			return true, nil
		}
		return false, nil
	})
	checkErr("waiting for server to boot", err)

	*content += fmt.Sprintf("Created a server!\n\n")
	*content += fmt.Sprintf("| %-20s | %-36s | %-15s | %-8s |\n", "Name", "ID", "Password", "RAM (GB)")
	*content += fmt.Sprintf("|%s|%s|%s|%s|\n", hyphens(20), hyphens(36), hyphens(15), hyphens(8))
	*content += fmt.Sprintf("| %-20s | %-36s | %-15s | %-8d |\n", name, server.ID, server.AdminPass, 1)

	return ip
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

	ensureMethod(r, "POST")

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	w.Header().Set("Access-Control-Expose-Headers", "Server-IP")

	client := setupClients()["compute"]
	content := ""

	ip := createServer(client, &content, getName(r))
	w.Header().Set("Server-IP", ip)

	fmt.Fprintf(w, content)
}
