package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/images"
)

const imCachePath = "/tmp/sandbox_images"

func listImages(w http.ResponseWriter, r *http.Request) error {
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

	if err != nil {
		return err
	}

	ioutil.WriteFile(imCachePath, []byte(content), 0644)

	fmt.Fprintf(w, content)

	return nil
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://104.130.226.169:80")

	if fileExists(imCachePath) {
		content, err := ioutil.ReadFile(imCachePath)
		if string(content) != "" && err == nil {
			fmt.Fprintf(w, string(content))
			return
		}
	}

	err := listImages(w, r)

	// Catch 401 errors
	if casted, ok := err.(*perigee.UnexpectedResponseCodeError); ok && casted.Actual == 401 {
		if fileExists(cachePath) {
			os.Remove(cachePath)
			listImages(w, r)
		}
	} else {
		checkErr("listing images", err)
	}
}
