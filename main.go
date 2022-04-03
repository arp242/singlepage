package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"zgo.at/singlepage/singlepage"
)

const help = `
Bundle local and remote assets in a HTML file.

The -local, -remote, and -minify accept a comma-separated list of asset types;
the default is to include all the supported types. Pass an empty string to
disable the feature (e.g. -remote '').

Assets are looked up relative to the path in -root which may be a remote path
(e.g. http://example.com), in which case all resources are fetched relative to
that domain (and are treated as external).

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
		_, _ = fmt.Fprintf(os.Stderr, "singlepage: error: "+err.Error()+"\n")
		os.Exit(1)
	}

	fmt.Println(html)
}

func start() (string, error) {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "usage: singlepage [flags] file.html\n")
		_, _ = fmt.Fprint(os.Stderr, help)
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}

	root := flag.String("root", "", "look up assets relative to this path")
	local := flag.String("local", "css,js,img", "")
	remote := flag.String("remote", "css,js,img", "")
	minify := flag.String("minify", "css,js,html", "")
	strict := flag.Bool("strict", false,
		"fail on lookup or parse errors instead of leaving the content alone")
	flag.Parse()

	opts := singlepage.NewOptions(*root, *strict)
	err := opts.Commandline(*local, *remote, *minify)
	if err != nil {
		flag.Usage()
		return "", err
	}

	paths := flag.Args()
	if len(paths) != 1 {
		flag.Usage()
		return "", fmt.Errorf("must specify the path to exactly one HTML file")
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
