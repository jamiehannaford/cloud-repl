package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	th "github.com/rackspace/gophercloud/testhelper"
)

func authJSON(url string) string {
	return fmt.Sprintf(`
  {
    "access": {
      "serviceCatalog": [
      {
        "endpoints": [
        {
          "publicURL": "%s",
          "region": "IAD",
          "tenantId": "123456"
        }
        ],
        "name": "cloudLoadBalancers",
        "type": "rax:load-balancer"
        },
        {
          "endpoints": [
          {
            "publicURL": "%s",
            "region": "IAD",
            "tenantId": "123456",
            "versionId": "2",
            "versionInfo": "%s",
            "versionList": "%s"
          }
          ],
          "name": "cloudServersOpenStack",
          "type": "compute"
        }
        ],
        "token": {
          "RAX-AUTH:authenticatedBy": ["APIKEY"],
          "expires": "2013-11-13T10:49:29.409Z",
          "id": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
          "tenant": {"id": "123456","name": "123456"}
        }
      }
    }
`, url, url, url, url)
}

func serverJSON(id, pwd string) string {
	return fmt.Sprintf(`
{
  "server": {
    "OS-DCF:diskConfig": "AUTO",
    "id": "%s",
    "adminPass": "%s"
  }
}
`, id, pwd)
}

func assertNoErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Error was raised: %s", err)
	}
}

func trim(str string) string {
	return strings.Trim(str, "\n ")
}

func TestServerHandle(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest("POST", "/create_server", strings.NewReader(`{"name": "server_1"}`))
	th.AssertNoErr(t, err)
	req.Header.Set("User-Prefix", "SOME_PREFIX")

	res := httptest.NewRecorder()

	// Ensure that gophercloud hits our test server
	os.Setenv("RS_AUTH_URL", ts.URL+"/v2.0")

	// First we mock the auth call
	mux.HandleFunc("/v2.0/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, authJSON(ts.URL))
	})

	// Now create the server
	mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
		th.TestJSONRequest(t, r, `
{
	"server": {
		"name": "SOME_PREFIX_server_1",
		"OS-DCF:diskConfig": "MANUAL",
		"flavorRef": "general1-1",
		"imageRef": "a3da5530-71c6-4405-b64f-fd2da99d303c",
		"key_name": ""
	}
}
`)
		fmt.Fprintln(w, serverJSON("foo_id", "password"))
	})

	// Final request is to check the server has been built
	mux.HandleFunc("/servers/foo_id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, `{"server": {"status": "ACTIVE"}}`)
	})

	handleServerCreate(res, req)

	expected := trim(`
| Name                 | ID                                   | Password        |
|----------------------|--------------------------------------|-----------------|
| server_1             | foo_id                               | password        |
`)

	actual := trim(res.Body.String())

	th.AssertEquals(t, expected, actual)
}
