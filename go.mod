module zgo.at/singlepage

go 1.13

require (
	// TODO: goquery 1.8.0 escapes things inside <style> tags.
	//github.com/PuerkitoBio/goquery v1.8.0
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/tdewolff/minify/v2 v2.10.0
	github.com/tdewolff/parse/v2 v2.5.27
	zgo.at/zli v0.0.0-20220403205301-99207e5ec503
	zgo.at/zstd v0.0.0-20220306174247-aa79e904bd64
)

require github.com/andybalholm/cascadia v1.3.1 // indirect
