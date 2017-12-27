package singlepage

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/css"
)

// Replace <link rel="stylesheet" href="/_static/style.css"> with
// <style>..</style>
func replaceCSSLinks(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalCSS && !opts.RemoteCSS {
		return nil
	}

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
		f, err = readPath(path)
		if err != nil {
			return false
		}

		// Replace @imports
		var out string
		out, err = replaceCSSURLs(string(f))
		if err != nil {
			return false
		}

		if opts.MinifyCSS {
			out, err = minifier.String("css", out)
			if err != nil {
				return false
			}
		}

		s.AfterHtml("<style>" + out + "</style>")
		s.Remove()
		return true
	})
	return err
}

// Replace @import "path"; and url("..")
func replaceCSSImports(doc *goquery.Document, opts Options) (err error) {
	if !opts.LocalCSS && !opts.RemoteCSS {
		return nil
	}

	doc.Find("style").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var n string
		n, err = replaceCSSURLs(s.Text())
		if err != nil {
			return false
		}
		s.SetText(n)
		return true
	})
	return err
}

func replaceCSSURLs(s string) (string, error) {
	l := css.NewLexer(strings.NewReader(s))
	var out []byte
	for {
		tt, text := l.Next()
		switch {

		case tt == css.ErrorToken:
			if l.Err() == io.EOF {
				return string(out), nil
			}
			return string(out), l.Err()

		// @import
		case tt == css.AtKeywordToken && string(text) == "@import":
			for {
				tt2, text2 := l.Next()
				if tt2 == css.SemicolonToken {
					break
				}
				if tt2 == css.ErrorToken {
					return "", l.Err()
				}

				var path string
				if tt2 == css.StringToken {
					path = strings.Trim(string(text2), `'"`)
				} else if tt2 == css.URLToken {
					path = string(text2)
					path = path[strings.Index(path, "(")+1 : strings.Index(path, ")")]
					path = strings.Trim(path, `'"`)
				} else {
					continue
				}

				if path != "" {
					b, err := readPath(path)
					if err != nil {
						return "", err
					}

					nest, err := replaceCSSURLs(string(b))
					if err != nil {
						return "", err
					}
					out = append(out, []byte(nest)...)
				}
			}

		// Images
		case tt == css.URLToken:
			path := string(text)
			path = path[strings.Index(path, "(")+1 : strings.Index(path, ")")]
			path = strings.Trim(path, `'"`)

			f, err := readPath(path)
			if err != nil {
				return "", err
			}
			m := mime.TypeByExtension(filepath.Ext(path))
			out = append(out, []byte(fmt.Sprintf("url(data:%v;base64,%v)",
				m, base64.StdEncoding.EncodeToString(f)))...)

		default:
			out = append(out, text...)
		}
	}
}
