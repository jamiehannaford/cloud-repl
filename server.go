package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/images"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

const (
	defaultImage  = "a3da5530-71c6-4405-b64f-fd2da99d303c" // Ubuntu 12.04 LTS (Precise Pangolin) (PVHVM)
	defaultFlavor = "general1-1"                           // 1 GB General Purpose v1
)

func checkErr(msg string, err error) {
	if err != nil {
		panic(fmt.Sprintf("An error occurred while %s: %s", msg, err.Error()))
	}
}

type clientList map[string]*gophercloud.ServiceClient

func setupClients() clientList {
	opts, err := rackspace.AuthOptionsFromEnv()
	checkErr("retrieving env vars", err)

	region := os.Getenv("RS_REGION")

	client, err := rackspace.AuthenticatedClient(opts)
	checkErr("authenticating", err)

	compute, err := rackspace.NewComputeV2(client, gophercloud.EndpointOpts{Region: region})
	checkErr("creating compute service", err)

	return clientList{
		"compute": compute,
	}
}

func catchPanic(w http.ResponseWriter) func() {
	return func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}
}

func hyphens(count int) string {
	return strings.Repeat("-", count+2)
}

func handleFlavors(w http.ResponseWriter, r *http.Request) {
	defer catchPanic(w)

	client := setupClients()["compute"]
	content := ""

	err := flavors.ListDetail(client, nil).EachPage(func(page pagination.Page) (bool, error) {
		fs, err := flavors.ExtractFlavors(page)
		checkErr("extracting flavors", err)

		content += fmt.Sprintf("| %-23s | %-8s | %-9s | %-4s |\n", "Name", "RAM (GB)", "Disk (GB)", "CPUs")
		content += fmt.Sprintf("|%s|%s|%s|%s|\n", hyphens(23), hyphens(8), hyphens(9), hyphens(4))

		for _, f := range fs {
			RAMGB := f.RAM / 1024
			if RAMGB < 1 {
				continue
			}
			content += fmt.Sprintf("| %-23s | %-8d | %-9d | %-4d |\n", f.Name, RAMGB, f.Disk, f.VCPUs)
		}

		return true, nil
	})
	checkErr("listing flavors", err)

	fmt.Fprintf(w, content)
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	defer catchPanic(w)

	client := setupClients()["compute"]
	content := ""

	err := images.ListDetail(client, nil).EachPage(func(page pagination.Page) (bool, error) {
		is, err := images.ExtractImages(page)
		checkErr("extracting images", err)

		var images sort.StringSlice
		for _, i := range is {
			images = append(images, i.Name)
		}

		images.Sort()

		maxLength := 50

		content += fmt.Sprintf("| %-"+strconv.Itoa(maxLength)+"s |\n", "Name")
		content += fmt.Sprintf("|%s|\n", hyphens(maxLength))

		for _, name := range images {
			if len(name) > maxLength {
				name = name[:maxLength-3] + "..."
			}
			content += fmt.Sprintf("| %-"+strconv.Itoa(maxLength)+"s |\n", name)
		}

		return true, nil
	})
	checkErr("listing images", err)

	fmt.Fprintf(w, content)
}

func randomStr(prefix string, n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return prefix + string(bytes)
}

func handleServerCreate(w http.ResponseWriter, r *http.Request) {
	defer catchPanic(w)

	client := setupClients()["compute"]
	content := ""

	name := randomStr("sandbox", 10)

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

	content += fmt.Sprintf("Server created!\n\n")

	content += fmt.Sprintf("| %-20s | %-36s | %-15s |\n", "Name", "ID", "Password")
	content += fmt.Sprintf("|%s|%s|%s|\n", hyphens(20), hyphens(36), hyphens(15))
	content += fmt.Sprintf("| %-20s | %-36s | %-15s | \n", name, server.ID, server.AdminPass)

	fmt.Fprintf(w, content)
}

func handleServerQuery(w http.ResponseWriter, r *http.Request) {
	defer catchPanic(w)
}

func handleLBQuery(w http.ResponseWriter, r *http.Request) {
	defer catchPanic(w)
}

func main() {
	http.HandleFunc("/images", handleImages)
	http.HandleFunc("/flavors", handleFlavors)
	http.HandleFunc("/create_server", handleServerCreate)
	http.HandleFunc("/query_server", handleServerQuery)
	http.HandleFunc("/query_lb", handleLBQuery)

	http.ListenAndServe(":8080", nil)
}
