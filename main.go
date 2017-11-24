package main // import "arp242.net/singlepage"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"arp242.net/singlepage/singlepage"
)

const help = `
Bundle local and remote assets in a HTML file.

The -local, -remote, and -minify accept a comma-separated list of asset types;
the default is to include all the supported types.

Local assets are looked up relative to the path in -root. The -root may be a
remote path (e.g. http://example.com), in which case all resources are fetched
relative to that domain (and are treated as external).

For remote assets only "http://", "https://", and "//" are supported.

Limitations:

- Only <img src=".."> are bundled; not CSS images (e.g. background-image: url(..)).
- Nested style sheets with @import are not bundled.
- fonts are not bundled.
- We should support 'minification' of images

Flags:

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: singlepage [flags] file.html\n")
		fmt.Fprintf(os.Stderr, help)
		flag.PrintDefaults()
		os.Exit(2)
	}

	var opts singlepage.Options
	var local, remote, minify string
	flag.StringVar(&opts.Root, "root", "./", "")
	flag.StringVar(&local, "local", "css,js,img", "")
	flag.StringVar(&remote, "remote", "css,js,img", "")
	flag.StringVar(&minify, "minify", "css,js,html", "")
	flag.Parse()

	for _, v := range strings.Split(local, ",") {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "css":
			opts.LocalCSS = true
		case "js":
			opts.LocalJS = true
		case "img":
			opts.LocalImg = true
		default:
			fmt.Fprintf(os.Stderr, "error: unknown value for -local: %#v\n\n", v)
			flag.Usage()
		}
	}
	for _, v := range strings.Split(remote, ",") {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "css":
			opts.RemoteCSS = true
		case "js":
			opts.RemoteJS = true
		case "img":
			opts.RemoteImg = true
		default:
			fmt.Fprintf(os.Stderr, "error: unknown value for -remote: %#v\n\n", v)
			flag.Usage()
		}
	}
	for _, v := range strings.Split(minify, ",") {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "css":
			opts.MinifyCSS = true
		case "js":
			opts.MinifyJS = true
		case "html":
			opts.MinifyHTML = true
		default:
			fmt.Fprintf(os.Stderr, "error: unknown value for -minify: %#v\n\n", v)
			flag.Usage()
		}
	}

	paths := flag.Args()
	if len(paths) != 1 {
		fmt.Fprintf(os.Stderr, "error: must specify the path to exactly one HTML file\n\n")
		flag.Usage()
	}

	b, err := ioutil.ReadFile(paths[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	h, err := singlepage.Bundle(string(b), opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(h)
}
