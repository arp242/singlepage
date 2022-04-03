Inline CSS, JavaScript, and images in a HTML file to distribute a stand-alone
HTML document without external dependencies.

You can download binaries from the [releases] page, or compile from source with
`go install zgo.at/singlepage/cmd/singlepage@latest`, which will put a binary in
`~/go/bin/`.

Run it with as `singlepage file.html > bundled.html` or `cat file.html |
singlepage > bundled.html`. There are a bunch of options; use `singlepage -help`
to see the full documentation.

Use the `zgo.at/singlepage` package if you want to integrate this in a Go
program. Also see the API docs: https://godocs.io/zgo.at/singlepage

It uses [tdewolff/minify] for minification, so please report bugs or other
questions there.

[tdewolff/minify]: https://github.com/tdewolff/minify
[releases]: https://github.com/arp242/singlepage/releases


Why would I want to use this?
-----------------------------
There are a few reasons:

- Sometimes distributing a single HTML document is easier; for example for
  rendered HTML documentation.

- It makes pages slightly faster to load if your CSS/JS assets are small(-ish);
  especially on slower connections.

- As a slightly less practical and more ideological point, I liked the web
  before it became this jumbled mess of obnoxious JavaScript and excessive CSS,
  and I like the concept of self-contained HTML documents.
