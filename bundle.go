package singlepage

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"zgo.at/zstd/zint"
)

const (
	_    zint.Bitflag16 = 0
	HTML zint.Bitflag16 = 1 << (iota - 1)
	CSS
	JS
	Image
	Font
)

// Options for Bundle().
type Options struct {
	Root   string
	Strict bool
	Quiet  bool
	Local  zint.Bitflag16
	Remote zint.Bitflag16
	Minify zint.Bitflag16
}

// Everything is an Options struct with everything enabled.
var Everything = Options{
	Local:  CSS | JS | Image,
	Remote: CSS | JS | Image,
	Minify: CSS | JS | Image,
}

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("css", css.Minify)
	minifier.AddFunc("html", html.Minify)
	minifier.AddFunc("js", js.Minify)
}

// NewOptions creates a new Options instance.
func NewOptions(root string, strict, quiet bool) Options {
	return Options{Root: root, Strict: strict, Quiet: quiet}
}

// Commandline modifies the Options from the format accepted in the commandline
// tool's flags.
func (opts *Options) Commandline(local, remote, minify []string) error {
	for _, v := range local {
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "":
			continue
		case "css":
			opts.Local |= CSS
		case "js", "javascript":
			opts.Local |= JS
		case "img", "image", "images":
			opts.Local |= Image
		case "font", "fonts":
			opts.Local |= Font
		default:
			return fmt.Errorf("unknown value for -local: %q", v)
		}
	}
	for _, v := range remote {
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "":
			continue
		case "css":
			opts.Remote |= CSS
		case "js", "javascript":
			opts.Remote |= JS
		case "img", "image", "images":
			opts.Remote |= Image
		case "font", "fonts":
			opts.Remote |= Font
		default:
			return fmt.Errorf("unknown value for -remote: %q", v)
		}
	}
	for _, v := range minify {
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "":
			continue
		case "css":
			opts.Minify |= CSS
		case "js", "javascript":
			opts.Minify |= JS
		case "html":
			opts.Minify |= HTML
		default:
			return fmt.Errorf("unknown value for -minify: %q", v)
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

	if err := minifyStyleTags(doc, opts); err != nil {
		return "", fmt.Errorf("minifyStyleTags: %w", err)
	}
	if err := replaceCSSLinks(doc, opts); err != nil {
		return "", fmt.Errorf("replaceCSSLinks: %w", err)
	}
	if err := replaceCSSImports(doc, opts); err != nil {
		return "", fmt.Errorf("replaceCSSImports: %w", err)
	}
	if err := replaceJS(doc, opts); err != nil {
		return "", fmt.Errorf("replaceJS: %w", err)
	}
	if err := replaceImg(doc, opts); err != nil {
		return "", fmt.Errorf("replaceImg: %w", err)
	}

	h, err := doc.Html()
	if err != nil {
		return "", err
	}
	if opts.Minify.Has(HTML) {
		return minifier.String("html", h)
	}
	return h, nil
}

func minifyStyleTags(doc *goquery.Document, opts Options) (err error) {
	if !opts.Minify.Has(CSS) {
		return nil
	}

	doc.Find(`style`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		var f string
		f, err = minifier.String("css", s.Text())
		if err != nil {
			return false
		}
		s.SetText(f)
		return true
	})

	return err
}

func replaceJS(doc *goquery.Document, opts Options) (err error) {
	if !opts.Local.Has(JS) && !opts.Remote.Has(JS) {
		return nil
	}

	var cont bool
	doc.Find(`script`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("src")
		if !ok {
			if !opts.Minify.Has(JS) {
				return true
			}

			var f string
			f, err = minifier.String("js", s.Text())
			if err != nil {
				return false
			}
			s.SetText(f)
			return true
		}
		path = opts.Root + path

		if isRemote(path) && !opts.Remote.Has(JS) {
			return true
		}
		if !isRemote(path) && !opts.Local.Has(JS) {
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

		if opts.Minify.Has(JS) {
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
	if !opts.Local.Has(Image) && !opts.Remote.Has(Image) {
		return nil
	}

	var cont bool
	doc.Find(`img`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		path, ok := s.Attr("src")
		if !ok {
			return true
		}
		path = opts.Root + path

		if strings.HasPrefix(path, "data:") {
			return true
		}

		if isRemote(path) && !opts.Remote.Has(Image) {
			return true
		}
		if !isRemote(path) && !opts.Local.Has(Image) {
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

		m := mime.TypeByExtension(filepath.Ext(path))
		if m == "" {
			cont, err = warn(opts, &ParseError{Path: path, Err: errors.New("could not find MIME type")})
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
