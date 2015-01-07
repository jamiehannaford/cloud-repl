package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rackspace/gophercloud/rackspace/lb/v1/lbs"
	n "github.com/rackspace/gophercloud/rackspace/lb/v1/nodes"
	"github.com/rackspace/gophercloud/rackspace/lb/v1/vips"
)

func getNameAndIPs(w http.ResponseWriter, r *http.Request) (string, []string) {
	decoder := json.NewDecoder(r.Body)

	body := struct {
		Name string   `json:"name"`
		IPs  []string `json:"server_ips"`
	}{}

	err := decoder.Decode(&body)
	checkErr("decoding create_lb request JSON", err)

	if body.Name == "" {
		body.Name = randomStr("", 10)
	}
	if len(body.IPs) == 0 {
		panic("No server IPs provided")
	}

	return body.Name, body.IPs
}

func handleLBCreate(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	ensureMethod(r, "POST")

	w.Header().Set("Access-Control-Allow-Origin", "http://104.130.226.169:80")

	name, ips := getNameAndIPs(w, r)
	nodes := []n.Node{}

	for _, ip := range ips {
		node := n.Node{Address: string(ip), Port: 80, Condition: "ENABLED"}
		nodes = append(nodes, node)
	}

	opts := &lbs.CreateOpts{
		Name:     randomStr("sandbox_", 10) + name,
		Nodes:    nodes,
		Protocol: "HTTPS",
		VIPs: []vips.VIP{
			vips.VIP{Type: vips.PUBLIC, Version: vips.IPV4},
		},
	}

	client := setupClients()["lb"]

	lb, err := lbs.Create(client, opts).Extract()
	checkErr("creating load balancer", err)

	nameLen := strconv.Itoa(len(name))
	if nameLen < 4 {
		nameLen = 4
	}

	content := fmt.Sprintf("Created a load balancer!\n\n")
	content += fmt.Sprintf("| %-"+nameLen+"s | %-10s | %-15s |\n", "Name", "ID", "IPv4")
	content += fmt.Sprintf("|%s|%s|%s|\n", hyphens(len(name)), hyphens(10), hyphens(15))
	content += fmt.Sprintf("| %-"+nameLen+"s | %-10d | %-15s |\n", name, lb.ID, lb.SourceAddrs.IPv4Public)

	fmt.Fprintf(w, content)
}
