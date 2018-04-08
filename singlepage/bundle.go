package singlepage // import "arp242.net/singlepage/singlepage"

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

// Options for Bundle().
type Options struct {
	Root                            string
	LocalCSS, LocalJS, LocalImg     bool
	RemoteCSS, RemoteJS, RemoteImg  bool
	MinifyCSS, MinifyJS, MinifyHTML bool
}

// Everything is an Options struct with everything enabled.
var Everything = Options{
	LocalCSS: true, LocalImg: true, LocalJS: true, MinifyCSS: true,
	MinifyHTML: true, MinifyJS: true, RemoteCSS: true, RemoteImg: true,
	RemoteJS: true}

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("css", css.Minify)
	minifier.AddFunc("html", html.Minify)
	minifier.AddFunc("js", js.Minify)
}

// NewOptions creates a new Options from the format accepted in the commandline
// tool's flags.
func NewOptions(root, local, remote, minify string) (Options, error) {
	opts := Options{Root: root}

	for _, v := range strings.Split(strings.ToLower(local), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case "css":
			opts.LocalCSS = true
		case "js":
			opts.LocalJS = true
		case "img":
			opts.LocalImg = true
		default:
			return opts, fmt.Errorf("unknown value for -local: %#v", v)
		}
	}
	for _, v := range strings.Split(strings.ToLower(remote), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case "css":
			opts.RemoteCSS = true
		case "js":
			opts.RemoteJS = true
		case "img":
			opts.RemoteImg = true
		default:
			return opts, fmt.Errorf("unknown value for -remote: %#v", v)
		}
	}
	for _, v := range strings.Split(strings.ToLower(minify), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case "css":
			opts.MinifyCSS = true
		case "js":
			opts.MinifyJS = true
		case "html":
			opts.MinifyHTML = true
		default:
			return opts, fmt.Errorf("unknown value for -minify: %#v", v)
		}
	}
	return opts, nil
}

// Bundle the resources in a HTML document according to the given options.
func Bundle(html []byte, opts Options) (string, error) {
	if opts.Root != "./" {
		opts.Root = strings.TrimRight(opts.Root, "/")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return "", err
	}

	if err := replaceCSSLinks(doc, opts); err != nil {
		return "", err
	}
	if err := replaceCSSImports(doc, opts); err != nil {
		return "", err
	}
	if err := replaceJS(doc, opts); err != nil {
		return "", err
	}
	if err := replaceImg(doc, opts); err != nil {
		return "", err
	}

	h, err := doc.Html()
	if err != nil {
		return "", err
	}
	if opts.MinifyHTML {
		return minifier.String("html", h)
	}
	return h, nil
}

func replaceJS(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalJS && !opts.RemoteJS {
		return nil
	}

	doc.Find(`script`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("src")
		if !ok {
			return true
		}
		path = opts.Root + path

		if isRemote(path) && !opts.RemoteJS {
			return true
		}
		if !isRemote(path) && !opts.LocalJS {
			return true
		}

		var f []byte
		f, err = readPath(path)
		if err != nil {
			return false
		}

		if opts.MinifyJS {
			f, err = minifier.Bytes("js", f)
			if err != nil {
				return false
			}
		}

		s.AfterHtml("<script>" + string(f) + "</script>")
		s.Remove()
		return true
	})

	return err
}

func replaceImg(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalImg && !opts.RemoteImg {
		return nil
	}

	doc.Find(`img`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("src")
		if !ok {
			return true
		}
		path = opts.Root + path

		if isRemote(path) && !opts.RemoteImg {
			return true
		}
		if !isRemote(path) && !opts.LocalImg {
			return true
		}

		var f []byte
		f, err = readPath(path)
		if err != nil {
			return false
		}

		m := mime.TypeByExtension(filepath.Ext(path))
		if m == "" {
			err = fmt.Errorf("could not find MIME type for %#v", path)
			return false
		}

		s.SetAttr("src", fmt.Sprintf("data:%v;base64,%v",
			m, base64.StdEncoding.EncodeToString(f)))
		return true
	})

	return err
}
