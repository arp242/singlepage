[![GoDoc](https://godoc.org/arp242.net/singlepage?status.svg)](https://godoc.org/arp242.net/singlepage)
[![Build Status](https://travis-ci.org/Carpetsmoker/singlepage.svg?branch=master)](https://travis-ci.org/Carpetsmoker/singlepage)
[![codecov](https://codecov.io/gh/Carpetsmoker/singlepage/branch/master/graph/badge.svg)](https://codecov.io/gh/Carpetsmoker/singlepage)
[![Go Report Card](https://goreportcard.com/badge/arp242.net/singlepage)](https://goreportcard.com/report/arp242.net/singlepage)

Bundle external assets in a HTML file to distribute a "stand-alone" HTML
document.

This program is written in Go. To install it, you'll need to have Go installed;
then install this with:

	go get arp242.net/singlepage

Which should install it to `~/go/bin/singlepage`.

Running it is as easy as `singlepage file.html > bundled.html`. There are a
bunch of options; use `singlepage -help` to see the full documentation.

Use the `arp242.net/singlepage/singlepage` package if you want to integrate this
in a Go program; [`godocgen`](https://github.com/Teamwork/godocgen) does this
for example.

It uses [tdewolff/minify](https://github.com/tdewolff/minify) for minification,
so please report bugs or other questions there.
