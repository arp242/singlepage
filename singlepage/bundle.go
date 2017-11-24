package singlepage // import "arp242.net/singlepage/singlepage"

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

// Options for Bundle()
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
	RemoteJS: true,
}

// Bundle given external resources in a HTML document.
func Bundle(html string, opts Options) (string, error) {
	opts.Root = strings.TrimRight(opts.Root, "/")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	if err := replaceCSS(doc, opts); err != nil {
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
		return minifyHTML(h)
	}
	return h, nil
}

// Report if a path is remote.
func isRemote(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "//")
}

func readFile(path string) ([]byte, error) {
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

func replaceCSS(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalCSS && !opts.RemoteCSS {
		return nil
	}

	// <link rel="stylesheet" href="/_static/style.css">
	doc.Find(`link[rel="stylesheet"]`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("href")
		if !ok {
			return true
		}
		path = opts.Root + path

		if isRemote(path) && !opts.RemoteCSS {
			return true
		}
		if !isRemote(path) && !opts.LocalCSS {
			return true
		}

		var f []byte
		f, err = readFile(path)
		if err != nil {
			return false
		}

		if opts.MinifyCSS {
			f, err = minifyCSS(f)
			if err != nil {
				return false
			}
		}

		s.AfterHtml("<style>" + string(f) + "</style>")
		s.Remove()
		return true
	})
	return err
}

func replaceJS(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalJS && !opts.RemoteJS {
		return nil
	}

	// <script src="/_static/godocs.js"></script>
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
		f, err = readFile(path)
		if err != nil {
			return false
		}

		if opts.MinifyJS {
			f, err = minifyJS(f)
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
		f, err = readFile(path)
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

func minifyCSS(s []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	return m.Bytes("text/css", s)
}

func minifyJS(s []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("application/javascript", js.Minify)
	return m.Bytes("application/javascript", s)
}

func minifyHTML(s string) (string, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	return m.String("text/html", s)
}
