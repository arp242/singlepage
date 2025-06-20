package singlepage

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// Replace <link rel="stylesheet" href="/_static/style.css"> with
// <style>..</style>
func replaceCSSLinks(doc *goquery.Document, opts Options) (err error) {
	if !opts.Local.Has(CSS) && !opts.Remote.Has(CSS) {
		return nil
	}

	var cont bool
	doc.Find(`link[rel="stylesheet"]`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("href")
		if !ok {
			return true
		}
		path = opts.Root + path

		if isRemote(path) && !opts.Remote.Has(CSS) {
			return true
		}
		if !isRemote(path) && !opts.Local.Has(CSS) {
			return true
		}

		var f []byte
		f, err = readPath(path)
		cont, err = warn(opts, err)
		if err != nil {
			return false
		}
		if !cont {
			return true
		}

		// Replace @imports
		var out string
		out, err = replaceCSSURLs(opts, string(f))
		if err != nil {
			err = fmt.Errorf("could not parse %v: %v", path, err)
			return false
		}

		if opts.Minify.Has(CSS) {
			out, err = minifier.String("css", out)
			if err != nil {
				err = fmt.Errorf("could not minify %v: %v", path, err)
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
	if !opts.Local.Has(CSS) && !opts.Remote.Has(CSS) {
		return nil
	}

	doc.Find("style").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var n string
		n, err = replaceCSSURLs(opts, s.Text())
		if err != nil {
			err = fmt.Errorf("could not parse inline style block %v: %v", i, err)
			return false
		}
		s.SetHtml(n)
		return true
	})
	return err
}

func replaceCSSURLs(opts Options, s string) (string, error) {
	l := css.NewLexer(parse.NewInputString(s))
	var out []byte
	var cont bool
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
					cont, err = warn(opts, err)
					if err != nil {
						return "", err
					}
					if !cont {
						continue
					}

					nest, err := replaceCSSURLs(opts, string(b))
					if err != nil {
						return "", fmt.Errorf("could not load nested CSS file %v: %v", path, err)
					}
					out = append(out, []byte(nest)...)
				}
			}

		// Images and fonts
		case tt == css.URLToken:
			path := string(text)
			path = path[strings.Index(path, "(")+1 : strings.Index(path, ")")]
			if strings.HasPrefix(path, "data:") {
				out = append(out, text...)
				continue
			}

			path = strings.Trim(path, `'"`)
			m := mime.TypeByExtension(filepath.Ext(path))
			if m == "" {
				warn(opts, fmt.Errorf("unknown MIME type for %q; skipping", path))
				out = append(out, text...)
				continue
			}

			remote := isRemote(path)
			if strings.HasPrefix(m, "image/") &&
				((remote && !opts.Remote.Has(Image)) || (!remote && !opts.Local.Has(Image))) {
				out = append(out, text...)
				continue
			} else if strings.HasPrefix(m, "font/") &&
				((remote && !opts.Remote.Has(Font)) || (!remote && !opts.Local.Has(Font))) {
				out = append(out, text...)
				continue
			}

			f, err := readPath(path)
			cont, err = warn(opts, err)
			if err != nil {
				return "", err
			}
			if !cont {
				continue
			}

			out = append(out, []byte(fmt.Sprintf("url(data:%v;base64,%v)",
				m, base64.StdEncoding.EncodeToString(f)))...)

		default:
			out = append(out, text...)
		}
	}
}
