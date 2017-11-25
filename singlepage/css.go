package singlepage

import (
	"bytes"
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
		f, err = readFile(path)
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

// TODO: don't make it strip all whitespace.
func replaceCSSURLs(s string) (string, error) {
	p := css.NewParser(bytes.NewBufferString(s), false)

	out := ""
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			if p.Err() == io.EOF {
				return out, nil
			}
			return "", p.Err()

		// @import
		case css.AtRuleGrammar:
			if string(data) != "@import" {
				out += string(data) + formatCSSGrammar(p, gt)
			}

			var path string
			for _, val := range p.Values() {
				if val.TokenType == css.StringToken {
					path = strings.Trim(string(val.Data), `'"`)
				} else if val.TokenType == css.URLToken {
					d := string(val.Data)
					path = strings.Trim(d[strings.Index(d, "("):], `'"()`)
				}
			}

			if path != "" {
				b, err := readFile(path)
				if err != nil {
					return "", err
				}

				nest, err := replaceCSSURLs(string(b))
				if err != nil {
					return "", err
				}
				out += nest
			}

		// decl: url(..)
		case css.DeclarationGrammar:
			d := string(data)
			switch d {
			// TODO: Support background shorthand.
			// TODO: support cursor, list-style
			case "background-image":
				foundURL := false
				for _, v := range p.Values() {
					if v.TokenType == css.URLToken {
						foundURL = true
						d2 := string(v.Data)
						path := strings.Trim(d2[strings.Index(d2, "("):], `'"()`)
						f, err := readFile(path)
						if err != nil {
							return "", err
						}
						m := mime.TypeByExtension(filepath.Ext(path))
						out += fmt.Sprintf("%v:url(data:%v;base64,%v);",
							d, m, base64.StdEncoding.EncodeToString(f))
						break
					}
				}

				if !foundURL {
					out += d + formatCSSGrammar(p, gt)
				}
			default:
				out += d + formatCSSGrammar(p, gt)
			}

		case css.BeginAtRuleGrammar, css.BeginRulesetGrammar:
			out += string(data) + formatCSSGrammar(p, gt)

		default:
			out += string(data)
		}
	}
}

func formatCSSGrammar(p *css.Parser, gt css.GrammarType) (out string) {
	if gt == css.DeclarationGrammar {
		out += ":"
	}
	for _, val := range p.Values() {
		out += string(val.Data)
	}
	if gt == css.BeginAtRuleGrammar || gt == css.BeginRulesetGrammar {
		out += "{"
	} else if gt == css.AtRuleGrammar || gt == css.DeclarationGrammar {
		out += ";"
	}
	return out
}
