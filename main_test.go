// +build testhttp

package main

import (
	"flag"
	"os"
	"testing"
)

// Some real-world tests.
func TestMain(t *testing.T) {
	for _, tt := range []string{"github", "wikipedia", "aeon"} {
		t.Run(tt, func(t *testing.T) {
			os.Args = []string{"singlepage", "./testdata/" + tt + ".html"}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			_, err := start()
			if err != nil {
				t.Errorf("%T: %[1]s", err)
			}
		})
	}
}
