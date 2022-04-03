package singlepage

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"zgo.at/zstd/ztest"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		root                  string
		local, remote, minify []string
		want                  Options
	}{
		{"./", []string{"css"}, []string{""}, []string{"CSS", "jS"}, Options{
			Root:   "./",
			Local:  CSS,
			Minify: CSS | JS,
		}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := NewOptions(tt.root, false, false)
			err := out.Commandline(tt.local, tt.remote, tt.minify)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tt.want, out) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"./bundle_test.go", "package singlepage"},
		{"//example.com", "<!doctype html>"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out, err := readPath(tt.in)
			if err != nil {
				t.Fatal(err)
			}

			o := string(bytes.Split(out, []byte("\n"))[0])
			if o != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, tt.want)
			}
		})
	}
}

func TestReplaceJS(t *testing.T) {
	tests := []struct {
		in, want string
		opts     Options
		wantErr  string
	}{
		{
			`<script src="./testdata/a.js"></script>`,
			`<script>var foo={t:!0}</script>`,
			Options{Local: JS, Minify: JS},
			"",
		},
		{
			`<script src="./testdata/a.js"></script>`,
			"<script>var foo = {\n\tt: true,\n};\n</script>",
			Options{Local: JS},
			"",
		},
		{
			`<script src="./testdata/a.js"></script>`,
			`<script src="./testdata/a.js"></script>`,
			Options{},
			"",
		},
		{
			`<script src="./testdata/nonexist.js"></script>`,
			`<script src="./testdata/nonexist.js"></script>`,
			Options{Local: JS},
			"",
		},
		{
			`<script src="./testdata/nonexist.js"></script>`,
			`<script src="./testdata/nonexist.js"></script>`,
			Options{Local: JS, Strict: true},
			"no such file or directory",
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

			err = replaceJS(doc, tt.opts)
			if !ztest.ErrorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nout:  %v\nwant: %v\n", err, tt.wantErr)
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

func TestReplaceImg(t *testing.T) {
	tests := []struct {
		in, want string
		opts     Options
		wantErr  string
	}{
		{
			`<img src="./testdata/a.png"/>`,
			`<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4QsYBTofXds9gQAAAAZiS0dEAP8A/wD/oL2nkwAAAAxJREFUCB1jkPvPAAACXAEebXgQcwAAAABJRU5ErkJggg=="/>`,
			Options{Local: Image},
			"",
		},
		{
			`<img src="./testdata/a.png"/>`,
			`<img src="./testdata/a.png"/>`,
			Options{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			tt.in = `<html><head></head><body>` + tt.in + `</body></html>`
			tt.want = `<html><head></head><body>` + tt.want + `</body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.in))
			if err != nil {
				t.Fatal(err)
			}

			err = replaceImg(doc, tt.opts)
			if !ztest.ErrorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nout:  %v\nwant: %v\n", err, tt.wantErr)
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

func TestBundle(t *testing.T) {
	tests := []struct {
		in, want []byte
		opts     Options
		wantErr  string
	}{
		{
			ztest.Read(t, "./testdata/a.html"),
			ztest.Read(t, "./testdata/a.min.html"),
			Options{Minify: HTML},
			"",
		},
		//{
		//	ztest.Read(t, "./testdata/a.html"),
		//	ztest.Read(t, "./testdata/a.html"),
		//	Options{},
		//},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			o, err := Bundle(tt.in, tt.opts)
			if !ztest.ErrorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nout:  %v\nwant: %v\n", err, tt.wantErr)
			}

			want := strings.TrimSpace(string(tt.want))
			o = strings.TrimSpace(o)
			if o != want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", o, want)
			}
		})
	}
}
