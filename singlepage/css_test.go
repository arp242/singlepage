package singlepage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestReplaceCSSLinks(t *testing.T) {
	cases := []struct {
		in, want string
		opts     Options
	}{
		{
			`<link rel="stylesheet" href="./bundle_test/a.css">`,
			`<style>div{display:none}</style>`,
			Options{LocalCSS: true, MinifyCSS: true},
		},
		{
			`<link rel="stylesheet" href="./bundle_test/a.css">`,
			"<style>div {\n\tdisplay: none;\n}\n</style>",
			Options{LocalCSS: true},
		},
		{
			`<link rel="stylesheet" href="./bundle_test/a.css">`,
			`<link rel="stylesheet" href="./bundle_test/a.css"/>`,
			Options{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			tc.in = `<html><head>` + tc.in + `</head><body></body></html>`
			tc.want = `<html><head>` + tc.want + `</head><body></body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.in))
			if err != nil {
				t.Fatal(err)
			}

			err = replaceCSSLinks(doc, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			h, err := doc.Html()
			if err != nil {
				t.Fatal(err)
			}

			o := string(h)
			if o != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, tc.want)
			}
		})
	}
}

// nolint: lll
func TestImportImports(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{
			`
			@import './bundle_test/a.css';
			@import url("./bundle_test/a.css") print;
			span { display: block; }`,
			"div{display:none;}div{display:none;}span{display:block;}",
		},
		{
			`span { background-image: url('bundle_test/a.png'); }`,
			`span{background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4QsYBTofXds9gQAAAAZiS0dEAP8A/wD/oL2nkwAAAAxJREFUCB1jkPvPAAACXAEebXgQcwAAAABJRU5ErkJggg==);}`,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := replaceCSSURLs(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if out != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tc.want)
			}
		})
	}
}
