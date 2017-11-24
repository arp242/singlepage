package singlepage

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/teamwork/test"
)

func TestIsRemote(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"./asd", false},
		{"http://./asd", true},
		{"//./asd", true},
		{"://./asd", false},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			out := isRemote(tc.in)
			if out != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tc.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"./bundle_test.go", "package singlepage"},
		{"//example.com", "<!doctype html>"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			out, err := readFile(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			o := string(bytes.Split(out, []byte("\n"))[0])
			if o != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, tc.want)
			}
		})
	}
}

func TestReplaceCSS(t *testing.T) {
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

			err = replaceCSS(doc, tc.opts)
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

func TestReplaceJS(t *testing.T) {
	cases := []struct {
		in, want string
		opts     Options
	}{
		{
			`<script src="./bundle_test/a.js"></script>`,
			`<script>var foo={t:true,};</script>`,
			Options{LocalJS: true, MinifyJS: true},
		},
		{
			`<script src="./bundle_test/a.js"></script>`,
			"<script>var foo = {\n\tt: true,\n};\n</script>",
			Options{LocalJS: true},
		},
		{
			`<script src="./bundle_test/a.js"></script>`,
			`<script src="./bundle_test/a.js"></script>`,
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

			err = replaceJS(doc, tc.opts)
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
func TestReplaceImg(t *testing.T) {
	cases := []struct {
		in, want string
		opts     Options
	}{
		{
			`<img src="./bundle_test/a.png"/>`,
			`<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4QsYBTofXds9gQAAAAZiS0dEAP8A/wD/oL2nkwAAAAxJREFUCB1jkPvPAAACXAEebXgQcwAAAABJRU5ErkJggg=="/>`,
			Options{LocalImg: true},
		},
		{
			`<img src="./bundle_test/a.png"/>`,
			`<img src="./bundle_test/a.png"/>`,
			Options{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			tc.in = `<html><head></head><body>` + tc.in + `</body></html>`
			tc.want = `<html><head></head><body>` + tc.want + `</body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.in))
			if err != nil {
				t.Fatal(err)
			}

			err = replaceImg(doc, tc.opts)
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

func TestBundle(t *testing.T) {
	cases := []struct {
		in, want []byte
		opts     Options
	}{
		{
			test.Read(t, "./bundle_test/a.html"),
			test.Read(t, "./bundle_test/a.min.html"),
			Options{MinifyHTML: true},
		},
		//{
		//	test.Read(t, "./bundle_test/a.html"),
		//	test.Read(t, "./bundle_test/a.html"),
		//	Options{},
		//},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			o, err := Bundle(string(tc.in), tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			want := strings.TrimSpace(string(tc.want))
			o = strings.TrimSpace(o)
			if o != want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, want)
			}
		})
	}
}
