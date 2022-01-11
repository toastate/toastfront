package builder

import (
	"errors"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
)

var CSSBuilderImportRegexp = regexp.MustCompile(`(?m)^@import "local:\/\/(.*)";$`)

type CSSBuilder struct {
	builder *Builder
	depth   int
	data    map[string]interface{}
}

func (cb *CSSBuilder) Init() error {
	tlogger.Debug("builder", "css", "msg", "init")
	return nil
}

func (cb *CSSBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return cb.IsCssFile(path, file)
}

func (cb *CSSBuilder) IsCssFile(path string, file fs.FileInfo) bool {
	return filepath.Ext(file.Name()) == ".css"
}

func (bb *CSSBuilder) Process(path string, file fs.FileInfo) error {
	tlogger.Debug("builder", "css", "msg", "processing", "file", path)

	f, err := bb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	of, err := os.OpenFile(filepath.Join(bb.builder.BuildDir, path), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		tlogger.Error("builder", "css", "msg", "output file creation", "file", path, "err", err)
		return err
	}
	defer of.Close()

	wr := filewriter.Writer("text/css", of)

	t, err := template.New(path).Delims(`"{{`, `}}"`).Parse(string(f))

	err = t.Execute(wr, bb.data)
	if err != nil {
		tlogger.Error("builder", "css", "msg", "templater", "file", path, "err", err)
		return err
	}

	return nil
}

func (bb *CSSBuilder) ProcessAsByte(path string, file fs.FileInfo) ([]byte, error) {
	if bb.depth > 5 {
		tlogger.Debug("builder", "css", "msg", "file error", "file", path, "err", "reached max recursion depth of 5, import loop ?")
		return nil, errors.New("Too deep")
	}
	f, err := ioutil.ReadFile(filepath.Join(bb.builder.SrcDir, path))
	if err != nil {
		tlogger.Error("builder", "css", "msg", "file error", "file", path, "err", err)
		return nil, err
	}

	f = replaceWindowsCarriageReturn(f)

	f = CSSBuilderImportRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		p := string(CSSBuilderImportRegexp.FindSubmatch(match)[1])

		p = strings.ReplaceAll(p, "/", string(os.PathSeparator))

		if p[0] == os.PathSeparator {
			p = p[1:]
		} else {
			p = filepath.Join(filepath.Dir(path), p)
		}

		fileData, err := os.Stat(filepath.Join(bb.builder.SrcDir, p))
		if err != nil {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}

		if !bb.IsCssFile(p, fileData) {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", "import types mismatched")
			return []byte{'\n'}
		}

		nestedBB := &CSSBuilder{
			builder: bb.builder,
			depth:   bb.depth + 1,
			data:    bb.data,
		}

		if _, ok := bb.builder.FileDeps[p]; ok {
			bb.builder.FileDeps[p][path] = struct{}{}
		} else {
			bb.builder.FileDeps[p] = map[string]struct{}{path: {}}
		}

		c, err := nestedBB.ProcessAsByte(p, fileData)
		if err != nil {
			tlogger.Error("builder", "css", "msg", "file error process", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}
		if c[len(c)-1] != '\n' {
			c = append(c, '\n')
		}

		return c
	})

	return f, nil
}
