package dyndns

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Urls contains a set of mirrors in which a
// raw IP string can be retrieved. It is exported
// for the intent of modification.
var (
	Urls   = []string{"https://api.ipify.org"}
	v6Urls = []string{"https://api6.ipify.org"}
)

func tryMirror(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer dclose(resp.Body)
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func tryMirror6(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer dclose(resp.Body)
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

// Closing connections and handling the errors
func dclose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}

// GetExternalIP retrieves the external facing IP Address.
// If multiple mirrors are provided in Urls,
// it will try each one (in order), should
// preceding mirrors fail.
func GetExternalIP() (string, error) {
	for _, url := range Urls {
		resp, err := tryMirror(url)
		if err == nil {
			return resp, err
		} else {
			return "", err
		}
	}

	return "", errors.New("Could not retreive external IPv4")
}

func GetExternalIPv6() (string, error) {
	for _, url := range v6Urls {
		resp, err := tryMirror6(url)
		if err != nil {
			return "", err
		} else if !strings.Contains(resp, ":") {
			return "", errors.New("IPv6 address received was not valid")
		} else {
			return resp, nil
		}
	}

	return "", errors.New("Could not retreive external IPv6")
}
