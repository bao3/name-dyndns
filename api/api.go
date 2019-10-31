// Package api provides a basic interface for dealing
// with Name.com DNS API's.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/razoralpha/name-dyndns/log"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	productionURL = "https://api.name.com/v4/"
	devURL        = "https://api.dev.name.com/v4/"
)

// API Contains details required to access the Name.com API.
type API struct {
	baseURL  string
	username string
	token    string
}

// DNSRecord contains information about a Name.com DNS record.
type DNSRecord struct {
	RecordID   int    `json:"id"`
	DomainName string `json:"domainName"`
	Host       string `json:"host"`
	FQDN       string `json:"fqdn"`
	Type       string `json:"type"`
	Answer     string `json:"answer"`
	TTL        int    `json:"ttl"`
}

// NewNameAPI constructs a new Name.com API. If dev is true, then
// the API uses the development API, instead of the production API.
func NewNameAPI(username, token string, dev bool) API {
	a := API{username: username, token: token}

	if dev {
		a.baseURL = devURL
	} else {
		a.baseURL = productionURL
	}

	return a
}

// NewAPIFromConfig constructs a new Name.com API from a configuration.
func NewAPIFromConfig(c Config) API {
	return NewNameAPI(c.Username, c.Token, c.Dev)
}

func (api API) performRequest(method, url string, body io.Reader) (response []byte, err error) {
	var client http.Client
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Logger.Printf("Error building HTTP request: %s", err)
		return nil, err
	}

	req.SetBasicAuth(api.username, api.token)

	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Printf("Error making HTTP request %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Update DNS records instead of delete create otherwise some error might completely remove your entries
func (api API) UpdateDNSRecord(record DNSRecord) error {
	// We need to transform name -> hostname for JSON.
	var body struct {
		Host   string `json:"host"`
		Type   string `json:"type"`
		Answer string `json:"answer"`
		TTL    int    `json:"ttl"`
	}

	body.Host = record.Host
	body.Type = record.Type
	body.Answer = record.Answer
	body.TTL = record.TTL

	b, jsonerr := json.Marshal(body)
	if jsonerr != nil {
		return jsonerr
	}

	_, apierr := api.performRequest(
		"PUT",
		fmt.Sprintf("%s%s%s%s%d", api.baseURL, "domains/", record.DomainName, "/records/", record.RecordID),
		bytes.NewBuffer(b),
	)
	if apierr != nil {
		log.Logger.Printf("Error in Update (PUT) request: %s", apierr)
		return apierr
	}

	return nil
}

// GetDNSRecords returns a slice of DNS records associated with a given domain.
func (api API) GetDNSRecords(domain string) (records []DNSRecord, err error) {
	resp, err := api.performRequest(
		"GET",
		fmt.Sprintf("%s%s%s%s", api.baseURL, "domains/", domain, "/records"),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var result struct {
		Records []DNSRecord
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return result.Records, nil
}
