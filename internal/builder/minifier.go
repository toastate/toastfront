package builder

import (
	"io"
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

type FileWriter interface {
	Writer(string, io.WriteCloser) io.WriteCloser
}

var filewriter FileWriter

type TDMinifier struct {
	Minifier *minify.M
}

func (m *TDMinifier) Writer(mediatype string, out io.WriteCloser) io.WriteCloser {
	return m.Minifier.Writer(mediatype, out)
}

type NOOPMinifier struct {
}

func (m *NOOPMinifier) Writer(mediatype string, out io.WriteCloser) io.WriteCloser {
	return out
}

func init() {
	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	})
	minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	filewriter = &TDMinifier{
		Minifier: minifier,
	}
}
