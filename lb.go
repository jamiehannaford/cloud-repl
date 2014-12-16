package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rackspace/gophercloud/rackspace/lb/v1/lbs"
	n "github.com/rackspace/gophercloud/rackspace/lb/v1/nodes"
	"github.com/rackspace/gophercloud/rackspace/lb/v1/vips"
)

func getIPs(w http.ResponseWriter, r *http.Request) []string {
	decoder := json.NewDecoder(r.Body)

	body := struct {
		IPs []string `json:"server_ips"`
	}{}

	err := decoder.Decode(&body)
	checkErr("decoding create_lb request JSON", err)

	if len(body.IPs) == 0 {
		panic("No server IPs provided")
	}

	return body.IPs
}

func ensureMethod(r *http.Request, expected string) {
	if r.Method != expected {
		panic(fmt.Sprintf("%s is not an expected method", r.Method))
	}
}

func handleLBCreate(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, fmt.Sprintf("%s", r))
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")

	ensureMethod(r, "POST")

	ips := getIPs(w, r)
	nodes := []n.Node{}

	for _, ip := range ips {
		node := n.Node{Address: string(ip), Port: 80, Condition: "ENABLED"}
		nodes = append(nodes, node)
	}

	opts := &lbs.CreateOpts{
		Name:     randomStr("sandbox_", 10),
		Nodes:    nodes,
		Protocol: "HTTPS",
		VIPs: []vips.VIP{
			vips.VIP{Type: vips.PUBLIC, Version: vips.IPV4},
		},
	}

	client := setupClients()["lb"]

	lb, err := lbs.Create(client, opts).Extract()
	checkErr("creating load balancer", err)

	content := fmt.Sprintf("Created a load balancer!\n\n")
	content += fmt.Sprintf("| %-20s | %-10s | %-10s |\n", "Name", "ID", "IPv4")
	content += fmt.Sprintf("|%s|%s|%s|\n", hyphens(20), hyphens(10), hyphens(10))
	content += fmt.Sprintf("| %-20s | %-10d | %-10s |\n", lb.Name, lb.ID, lb.SourceAddrs.IPv4Public)

	fmt.Fprintf(w, content)
	w.Header().Set("Blah", "*")
}
