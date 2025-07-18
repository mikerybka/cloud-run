package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mikerybka/util"
)

func main() {
	createCloudVM()
	// installGo()
	// installCmd()
	// createSystemdUnit()
	// start()
}

const hetznerAPI = "https://api.hetzner.cloud/v1/servers"

type CreateServerRequest struct {
	Name             string   `json:"name"`
	ServerType       string   `json:"server_type"`
	Image            string   `json:"image"`
	Location         string   `json:"location,omitempty"` // or "datacenter"
	SSHKeys          []string `json:"ssh_keys,omitempty"` // name or ID
	StartAfterCreate bool     `json:"start_after_create"`
	UserData         string   `json:"user_data,omitempty"`
}

type ServerResponse struct {
	Server struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"server"`
}

func createCloudVM() {
	token := os.Getenv("HETZNER_TOKEN")
	if token == "" {
		fmt.Println("Set HETZNER_TOKEN environment variable")
		return
	}

	path := os.Args[1]

	name := util.RandomString(8, "abcdefghijklmnopqrstuvwxyz")

	cloudInit := fmt.Sprintf(`#cloud-config
runcmd:
  - apt update && apt install -y golang-go
  - go install github.com/mikerybka/install-go@latest
  - /usr/local/go/bin/go install %s@latest
  - /usr/local/go/bin/go install github.com/mikerybka/create-system-service@latest
  - /root/go/bin/create-system-service /root/go/bin/%s
`, path, filepath.Base(path))

	reqBody := CreateServerRequest{
		Name:       name,
		ServerType: "cpx11",
		Image:      "debian-12",
		Location:   "ash", // Ashburn, VA
		SSHKeys: []string{
			"laptop",
			"android",
			"beelink",
			"beelink2",
		},
		StartAfterCreate: true,
		UserData:         cloudInit,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", hetznerAPI, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	// io.Copy(os.Stdout, resp.Body)

	if resp.StatusCode != 201 {
		fmt.Println("Failed with status:", resp.Status)
		return
	}

	var out ServerResponse
	json.NewDecoder(resp.Body).Decode(&out)

	fmt.Println(out.Server.ID)
}
