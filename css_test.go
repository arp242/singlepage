package singlepage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestReplaceCSSLinks(t *testing.T) {
	tests := []struct {
		in, want string
		opts     Options
	}{
		{
			`<link rel="stylesheet" href="./testdata/a.css">`,
			`<style>div{display:none}</style>`,
			Options{Local: CSS, Minify: CSS},
		},
		{
			`<link rel="stylesheet" href="./testdata/a.css">`,
			"<style>div {\n\tdisplay: none;\n}\n</style>",
			Options{Local: CSS},
		},
		{
			`<link rel="stylesheet" href="./testdata/a.css">`,
			`<link rel="stylesheet" href="./testdata/a.css"/>`,
			Options{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			tt.in = `<html><head>` + tt.in + `</head><body></body></html>`
			tt.want = `<html><head>` + tt.want + `</head><body></body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.in))
			if err != nil {
				t.Fatal(err)
			}

			err = replaceCSSLinks(doc, tt.opts)
			if err != nil {
				t.Fatal(err)
			}

			h, err := doc.Html()
			if err != nil {
				t.Fatal(err)
			}

			o := string(h)
			if o != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, tt.want)
			}
		})
	}
}

func TestReplaceCSSImports(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{
			`<style>
				@import './testdata/a.css';
				span { display: block; }
			</style>`,
			"<style>\n\t\t\t\tdiv {\n\tdisplay: none;\n}\n\n\t\t\t\tspan { display: block; }\n\t\t\t</style>",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			tt.in = `<html><head>` + tt.in + `</head><body></body></html>`
			tt.want = `<html><head>` + tt.want + `</head><body></body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.in))
			if err != nil {
				t.Fatal(err)
			}

			err = replaceCSSImports(doc, Options{Local: CSS})
			if err != nil {
				t.Fatal(err)
			}
			h, err := doc.Html()
			if err != nil {
				t.Fatal(err)
			}

			o := string(h)
			if o != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, tt.want)
			}
		})
	}
}

func TestReplaceCSSURLs(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`span { display: block; }`, `span { display: block; }`},
		{`@import './testdata/a.css';`, "div {\n\tdisplay: none;\n}\n"},
		{`@import url("./testdata/a.css");`, "div {\n\tdisplay: none;\n}\n"},
		{`@import url("./testdata/a.css") print;`, "div {\n\tdisplay: none;\n}\n"},
		{
			`span { background-image: url('testdata/a.png'); }`,
			`span { background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4QsYBTofXds9gQAAAAZiS0dEAP8A/wD/oL2nkwAAAAxJREFUCB1jkPvPAAACXAEebXgQcwAAAABJRU5ErkJggg==); }`,
		},
		{
			`span { background-image: url(data:image/png;base64,iVBORw0KGgoAAA==); }`,
			`span { background-image: url(data:image/png;base64,iVBORw0KGgoAAA==); }`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := replaceCSSURLs(Options{Local: CSS | Image}, tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}
