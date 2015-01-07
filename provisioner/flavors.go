package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
)

const flCachePath = "/tmp/sandbox_flavors"

func listFlavors(w http.ResponseWriter, r *http.Request) error {
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
	if err != nil {
		return err
	}

	ioutil.WriteFile(flCachePath, []byte(content), 0644)

	fmt.Fprintf(w, content)

	return nil
}

func handleFlavors(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://104.130.226.169:80")

	if fileExists(flCachePath) {
		content, err := ioutil.ReadFile(flCachePath)
		if string(content) != "" && err == nil {
			fmt.Fprintf(w, string(content))
			return
		}
	}

	err := listFlavors(w, r)

	// Catch 401 errors
	if casted, ok := err.(*perigee.UnexpectedResponseCodeError); ok && casted.Actual == 401 {
		if fileExists(cachePath) {
			os.Remove(cachePath)
			listFlavors(w, r)
		}
	} else {
		checkErr("listing flavors", err)
	}
}
