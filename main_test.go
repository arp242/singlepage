// +build slowtest

package main

import (
	"flag"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	cases := []string{"github", "wikipedia", "aeon"}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			os.Args = []string{"singlepage", "./testdata/" + tc + ".html"}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			_, err := start()
			if err != nil {
				t.Error(err)
			}
		})
	}
}
