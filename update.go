package main

import (
	"encoding/json"
	"fmt"
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
	ip := ""
	t := time.NewTicker(1 * time.Second)
	for now := range t.C {
		fmt.Println("Checking DNS at ", now)
		var latestIp = getIp()
		if latestIp != ip {
			ip = latestIp
			updateDNS(ip)
		}
	}
}

func updateDNS(ip string) {
	records := getRecords()
	for _, e := range records.Data {
		if e.Type == "A" && e.Record == "home.pfista.io" {
			fmt.Println("Removing DNS A entry")
			removeRecord("home.pfista.io", "A", e.Value)
		}
		if e.Type == "CNAME" && e.Record == "home.pfista.io" {
			fmt.Println("Removing DNS CNAME entry")
			removeRecord("home.pfista.io", "CNAME", e.Value)
		}
	}
	addRecord("home.pfista.io", "A", ip)
}

func getIp() string {
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("IP: %s\n", body)
	return string(body)
}

func getRecords() DreamhostResponse {
	req, _ := http.NewRequest("GET", HOST, nil)
	q := req.URL.Query()
	q.Add("cmd", "dns-list_records")
	q.Add("unique_id", uuid())
	q.Add("format", "json")
	q.Add("key", KEY)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req.URL.String())

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
	fmt.Println(req.URL.String())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	res := DreamhostResponse{}
	json.Unmarshal([]byte(body), &res)

	fmt.Printf("data %s, response %s\n", res.Data, res.Result)

	if res.Result != "success" {
		fmt.Println(res.Result)
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
	fmt.Println(req.URL.String())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	res := DreamhostResponse{}
	json.Unmarshal([]byte(body), &res)

	fmt.Printf("data %s, response %s\n", res.Data, res.Result)

	if res.Result != "success" {
		fmt.Println(res.Result)
		fmt.Println(res.Data)
	}

}
