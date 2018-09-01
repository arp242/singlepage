package main // import "arp242.net/singlepage"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"arp242.net/singlepage/singlepage"
)

const help = `
Bundle local and remote assets in a HTML file.

The -local, -remote, and -minify accept a comma-separated list of asset types;
the default is to include all the supported types. Pass an empty string to
disable the feature (e.g. -remote '').

Local assets are looked up relative to the path in -root. The -root may be a
remote path (e.g. http://example.com), in which case all resources are fetched
relative to that domain (and are treated as external).

For remote assets only "http://", "https://", and "//" are supported; // is
treated as https://.

Limitations:

- Fonts are not bundled.
- We should support 'minification' of images (e.g. optipng).
- Everything is read in memory; you probably don't want to create very large
  documents with this (practically, this shouldn't be an issue for most sane
  documents, since browsers will start having problems after a few MB).

Flags:

`

func main() {
	html, err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	fmt.Println(html)
}

func start() (string, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: singlepage [flags] file.html\n")
		fmt.Fprintf(os.Stderr, help)
		flag.PrintDefaults()
		os.Exit(2)
	}

	var root, local, remote, minify string
	flag.StringVar(&root, "root", "", "")
	flag.StringVar(&local, "local", "css,js,img", "")
	flag.StringVar(&remote, "remote", "css,js,img", "")
	flag.StringVar(&minify, "minify", "css,js,html", "")
	flag.Parse()

	opts, err := singlepage.NewOptions(root, local, remote, minify)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		flag.Usage()
	}

	paths := flag.Args()
	if len(paths) != 1 {
		fmt.Fprintf(os.Stderr, "error: must specify the path to exactly one HTML file\n\n")
		flag.Usage()
	}

	b, err := ioutil.ReadFile(paths[0])
	if err != nil {
		return "", err
	}
	html, err := singlepage.Bundle(b, opts)
	if err != nil {
		return "", err
	}

	return html, nil
}
