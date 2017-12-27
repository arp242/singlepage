package singlepage

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Report if a path is remote.
func isRemote(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "//")
}

// Read a path, which may be either local or HTTP.
func readPath(path string) ([]byte, error) {
	if !isRemote(path) {
		if strings.HasPrefix(path, "/") {
			path = "." + path
		}
		return ioutil.ReadFile(path)
	}

	if strings.HasPrefix(path, "//") {
		path = "https:" + path
	}

	c := http.Client{Timeout: 5 * time.Second}
	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck
	return ioutil.ReadAll(resp.Body)
}
