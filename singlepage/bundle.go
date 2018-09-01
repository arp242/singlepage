package singlepage // import "arp242.net/singlepage/singlepage"

import (
	"bytes"
	"encoding/base64"
	"errors"
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
	Strict                          bool
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

const (
	optHTML = "html"
	optCSS  = "css"
	optJS   = "js"
	optImg  = "img"
)

func init() {
	minifier = minify.New()
	minifier.AddFunc(optCSS, css.Minify)
	minifier.AddFunc("html", html.Minify)
	minifier.AddFunc("js", js.Minify)
}

// NewOptions creates a new Options instance.
func NewOptions(root string, strict bool) Options {
	strictMode = strict
	return Options{Root: root, Strict: strict}
}

// Commandline modifies the Options from the format accepted in the commandline
// tool's flags.
func (opts *Options) Commandline(local, remote, minify string) error {
	for _, v := range strings.Split(strings.ToLower(local), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case optCSS:
			opts.LocalCSS = true
		case optJS:
			opts.LocalJS = true
		case optImg:
			opts.LocalImg = true
		default:
			return fmt.Errorf("unknown value for -local: %#v", v)
		}
	}
	for _, v := range strings.Split(strings.ToLower(remote), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case optCSS:
			opts.RemoteCSS = true
		case optJS:
			opts.RemoteJS = true
		case optImg:
			opts.RemoteImg = true
		default:
			return fmt.Errorf("unknown value for -remote: %#v", v)
		}
	}
	for _, v := range strings.Split(strings.ToLower(minify), ",") {
		switch strings.TrimSpace(v) {
		case "":
			continue
		case optCSS:
			opts.MinifyCSS = true
		case optJS:
			opts.MinifyJS = true
		case optHTML:
			opts.MinifyHTML = true
		default:
			return fmt.Errorf("unknown value for -minify: %#v", v)
		}
	}
	return nil
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

	var cont bool
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
		cont, err = warn(err)
		if err != nil {
			return false
		}
		if !cont {
			return true
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

	var cont bool
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
		cont, err = warn(err)
		if err != nil {
			return false
		}
		if !cont {
			return true
		}

		m := mime.TypeByExtension(filepath.Ext(path))
		if m == "" {
			cont, err = warn(&ParseError{Path: path, Err: errors.New("could not find MIME type")})
			if err != nil {
				return false
			}
			if !cont {
				return true
			}
		}

		s.SetAttr("src", fmt.Sprintf("data:%v;base64,%v",
			m, base64.StdEncoding.EncodeToString(f)))
		return true
	})

	return err
}
