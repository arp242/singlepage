package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"zgo.at/singlepage"
	"zgo.at/zli"
)

const usage = `usage: singlepage [flags] file.html

Bundle external assets in a HTML file to distribute a stand-alone HTML document.
https://github.com/arp242/singlepage

The -local, -remote, and -minify can be given more than once and/or accept a
comma-separated list of asset types; the default is to include all the supported
types. Pass an empty string to disable the feature (e.g. -remote '').

Flags:

    -h, -help      Show this help.

    -v, -version   Show version; add twice for detailed build info.

    -q, -quiet     Don't print warnings to stderr.

    -S, -strict    Fail on lookup or parse errors instead of leaving the content alone.

    -w, -write     Write the result to the input file instead of printing it.

    -r, -root      Assets are looked up relative to the path in -root, which may
                   be a remote path (e.g. http://example.com), in which case all
                   "//resources" are fetched relative to that domain (and are
                   treated as external).

    -l, -local     Filetypes to include from the local filesystem. Supports css,
                   js, img, and font.

    -r, -remote    Filetypes to include from remote sources. Only only
                   "http://", "https://", and "//" are supported; "//" is
                   treated as "https://". Suports css, js, img, and font.

    -m, -minify    Filetypes to minify. Support js, css, and html.
`

func fatal(err error) {
	if err == nil {
		return
	}

	zli.Errorf(err)
	fmt.Print("\n", usage)
	zli.Exit(1)
}

func main() {
	f := zli.NewFlags(os.Args)
	var (
		help     = f.Bool(false, "h", "help")
		versionF = f.IntCounter(0, "v", "version")
		quiet    = f.Bool(false, "q", "quiet")
		strict   = f.Bool(false, "S", "strict")
		write    = f.Bool(false, "w", "write")
		root     = f.String("", "r", "root", "")
		local    = f.StringList([]string{"css,js,img"}, "l", "local")
		remote   = f.StringList([]string{"css,js,img"}, "r", "remote")
		minify   = f.StringList([]string{"css,js,html"}, "m", "minify")
	)
	fatal(f.Parse())

	if help.Bool() {
		fmt.Print(usage)
		return
	}

	if versionF.Int() > 0 {
		zli.PrintVersion(versionF.Int() > 1)
		return
	}

	opts := singlepage.NewOptions(root.String(), strict.Bool(), quiet.Bool())
	err := opts.Commandline(local.StringsSplit(","), remote.StringsSplit(","), minify.StringsSplit(","))
	fatal(err)

	path := f.Shift()
	if path == "" && write.Bool() {
		fatal(errors.New("cannot use -write when reading from stdin"))
	}

	fp, err := zli.InputOrFile(path, quiet.Bool())
	fatal(err)
	defer fp.Close()

	b, err := io.ReadAll(fp)
	fatal(err)

	html, err := singlepage.Bundle(b, opts)
	fatal(err)

	if write.Bool() {
		fp.Close()
		fatal(os.WriteFile(path, []byte(html), 0644))
	} else {
		fmt.Println(html)
	}
}
