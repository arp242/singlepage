package singlepage

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/teamwork/test"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		root, local, remote, minify string
		want                        Options
	}{
		{"./", "css,", "", "CSS,jS", Options{
			Root:      "./",
			LocalCSS:  true,
			MinifyCSS: true,
			MinifyJS:  true,
		}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := NewOptions(tt.root, false)
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
			`<script>var foo={t:true,};</script>`,
			Options{LocalJS: true, MinifyJS: true},
			"",
		},
		{
			`<script src="./testdata/a.js"></script>`,
			"<script>var foo = {\n\tt: true,\n};\n</script>",
			Options{LocalJS: true},
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
			Options{LocalJS: true},
			"",
		},
		{
			`<script src="./testdata/nonexist.js"></script>`,
			`<script src="./testdata/nonexist.js"></script>`,
			Options{LocalJS: true, Strict: true},
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

			strictMode = tt.opts.Strict
			err = replaceJS(doc, tt.opts)
			if !test.ErrorContains(err, tt.wantErr) {
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

// nolint: lll
func TestReplaceImg(t *testing.T) {
	tests := []struct {
		in, want string
		opts     Options
		wantErr  string
	}{
		{
			`<img src="./testdata/a.png"/>`,
			`<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4QsYBTofXds9gQAAAAZiS0dEAP8A/wD/oL2nkwAAAAxJREFUCB1jkPvPAAACXAEebXgQcwAAAABJRU5ErkJggg=="/>`,
			Options{LocalImg: true},
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

			strictMode = tt.opts.Strict
			err = replaceImg(doc, tt.opts)
			if !test.ErrorContains(err, tt.wantErr) {
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
			test.Read(t, "./testdata/a.html"),
			test.Read(t, "./testdata/a.min.html"),
			Options{MinifyHTML: true},
			"",
		},
		//{
		//	test.Read(t, "./testdata/a.html"),
		//	test.Read(t, "./testdata/a.html"),
		//	Options{},
		//},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			strictMode = tt.opts.Strict
			o, err := Bundle(tt.in, tt.opts)
			if !test.ErrorContains(err, tt.wantErr) {
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
