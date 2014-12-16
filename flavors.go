package main

import (
	"fmt"
	"net/http"

	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
)

func handleFlavors(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")

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
