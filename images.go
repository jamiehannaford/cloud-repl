package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/images"
)

func handleImages(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")

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
