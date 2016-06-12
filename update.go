package main

import (
	"encoding/json"
	"fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const HOST = "https://api.dreamhost.com/"

var KEY = os.Getenv("DREAMHOST_DNS_API_KEY")

func main() {
	usage := `Dreamhost Dynamic DNS Updater.

Usage:
  dreamhost-ddns <record>

Options:
  -h --help     Show this screen.
  --version     Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "0.1.0", false)

	record, _ := arguments["<record>"].(string)
	dnstype := "A"

	// Get the current value for the dns record
	var ip string
	records := getRecords()
	for _, e := range records.Data {
		if e.Type == dnstype && e.Record == record {
			ip = e.Value
		}
	}

	// If the ip address is different than the current dns record, update it
	t := time.NewTicker(30 * time.Minute)
	for range t.C {
		var latestIp = getIp()
		if latestIp != ip {
			ip = latestIp
			updateDNS(record, dnstype, ip)
		}
	}
}

func getIp() string {
	resp, err := http.Get("https://ipv4.icanhazip.com")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return strings.TrimSpace(string(body))
}

func uuid() string {
	output, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
	uuid := strings.TrimSpace(string(output))
	return uuid
}

type DreamhostResponse struct {
	Data   []DnsRecord `json:"data"`
	Result string      `json:"result"`
}

type DnsRecord struct {
	AccountId string `json:"account_id"`
	Comment   string `json:"comment"`
	Editable  string `json:"editable"`
	Record    string `json:"record"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Zone      string `json:"zone"`
}

func updateDNS(record, dnstype, value string) {
	records := getRecords()
	var update = false
	for _, e := range records.Data {
		if e.Type == dnstype && e.Record == record && e.Value != value {
			removeRecord(record, dnstype, e.Value)
			update = true
		}
	}
	if update {
		addRecord(record, dnstype, value)
	}
}

func getRecords() DreamhostResponse {
	req, _ := http.NewRequest("GET", HOST, nil)
	q := req.URL.Query()
	q.Add("cmd", "dns-list_records")
	q.Add("unique_id", uuid())
	q.Add("format", "json")
	q.Add("key", KEY)
	req.URL.RawQuery = q.Encode()
	//fmt.Println(req.URL.String())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	res := DreamhostResponse{}
	json.Unmarshal([]byte(body), &res)
	return res
}

func addRecord(record, dnstype, value string) {
	req, _ := http.NewRequest("GET", HOST, nil)
	q := req.URL.Query()
	q.Add("cmd", "dns-add_record")
	q.Add("unique_id", uuid())
	q.Add("format", "json")
	q.Add("key", KEY)

	q.Add("record", record)
	q.Add("type", dnstype)
	q.Add("value", value)

	req.URL.RawQuery = q.Encode()
	//fmt.Println(req.URL.String())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	res := DreamhostResponse{}
	json.Unmarshal([]byte(body), &res)

	if res.Result == "success" {
		fmt.Printf("Successfully added dns record: %s, type: %s, value: %s\n", record, dnstype, value)
	} else {
		fmt.Printf("Error adding dns record: %s, type: %s, value: %s\n", record, dnstype, value)
		fmt.Println(res.Data)
	}
}

func removeRecord(record, dnstype, value string) {
	req, _ := http.NewRequest("GET", HOST, nil)
	q := req.URL.Query()
	q.Add("cmd", "dns-remove_record")
	q.Add("unique_id", uuid())
	q.Add("format", "json")
	q.Add("key", KEY)

	q.Add("record", record)
	q.Add("type", dnstype)
	q.Add("value", value)

	req.URL.RawQuery = q.Encode()
	//fmt.Println(req.URL.String())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	res := DreamhostResponse{}
	json.Unmarshal([]byte(body), &res)

	if res.Result == "success" {
		fmt.Printf("Successfully removed dns record: %s, type: %s, value: %s\n", record, dnstype, value)
	} else {
		fmt.Printf("Error removing dns record: %s, type: %s, value: %s\n", record, dnstype, value)
		fmt.Println(res.Data)
	}
}
