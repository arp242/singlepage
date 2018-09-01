package singlepage

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/teamwork/utils/stringutil"
)

// LookupError is used when we can't look up a resource. This may be a non-fatal
// error.
type LookupError struct {
	Path string
	Err  error
}

func (e *LookupError) Error() string { return e.Err.Error() }

// ParseError indicates there was a parsing failure. This may be a non-fatal
// error.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string { return e.Err.Error() }

// Report if a path is remote.
func isRemote(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "//")
}

var strictMode = false

// warn about an error.
func warn(err error) (bool, error) {
	switch err.(type) {

	case nil:
		return true, nil

	case *LookupError, *ParseError:
		if strictMode {
			return false, err
		}
		_, _ = fmt.Fprintf(os.Stderr, "singlepage: warning: %s\n", err)
		return false, nil

	default:
		return false, err
	}

}

// Read a path, which may be either local or HTTP.
func readPath(path string) ([]byte, error) {
	if !isRemote(path) {
		if strings.HasPrefix(path, "/") {
			path = "." + path
		}
		d, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, &LookupError{
				Path: path,
				Err:  err,
			}
		}
		return d, nil
	}

	if strings.HasPrefix(path, "//") {
		path = "https:" + path
	}

	c := http.Client{Timeout: 5 * time.Second}
	resp, err := c.Get(path)
	if err != nil {
		return nil, &LookupError{
			Path: path,
			Err:  err,
		}
	}
	defer resp.Body.Close() // nolint: errcheck

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &LookupError{
			Path: path,
			Err:  err,
		}
	}

	if resp.StatusCode != 200 {
		return nil, &LookupError{
			Path: path,
			Err: fmt.Errorf("%d %s: %s", resp.StatusCode, resp.Status,
				stringutil.Left(string(d), 100)),
		}
	}

	return d, nil
}
