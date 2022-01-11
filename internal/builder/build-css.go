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
	baseFolder string
	builder    *Builder
	depth      int
	data       map[string]interface{}
}

func (cb *CSSBuilder) Init() error {
	tlogger.Debug("builder", "css", "msg", "init")
	return nil
}

func (cb *CSSBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return cb.IsCssFile(path, file)
}

func (cb *CSSBuilder) IsCssFile(path string, file fs.FileInfo) bool {
	pathSplit := strings.Split(path, string(filepath.Separator))
	if pathSplit[0] != cb.baseFolder {
		return false
	}
	return filepath.Ext(file.Name()) == ".css"
}

func (cb *CSSBuilder) Process(path string, file fs.FileInfo) error {
	tlogger.Debug("builder", "css", "msg", "processing", "file", path)

	f, err := cb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	of, err := os.OpenFile(filepath.Join(cb.builder.BuildDir, path), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		tlogger.Error("builder", "css", "msg", "output file creation", "file", path, "err", err)
		return err
	}
	defer of.Close()

	wr := filewriter.Writer("text/css", of)

	t, err := template.New(path).Delims(`"{{`, `}}"`).Parse(string(f))

	err = t.Execute(wr, cb.data)
	if err != nil {
		tlogger.Error("builder", "css", "msg", "templater", "file", path, "err", err)
		return err
	}

	return nil
}

func (cb *CSSBuilder) ProcessAsByte(path string, file fs.FileInfo) ([]byte, error) {
	if cb.depth > 5 {
		tlogger.Debug("builder", "css", "msg", "file error", "file", path, "err", "reached max recursion depth of 5, import loop ?")
		return nil, errors.New("Too deep")
	}
	f, err := ioutil.ReadFile(filepath.Join(cb.builder.SrcDir, path))
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
		}

		p = filepath.Join(cb.baseFolder, p)

		fileData, err := os.Stat(filepath.Join(cb.builder.SrcDir, p))
		if err != nil {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}

		if !cb.IsCssFile(p, fileData) {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", "import types mismatched")
			return []byte{'\n'}
		}

		nestedCB := &CSSBuilder{
			baseFolder: cb.baseFolder,
			builder:    cb.builder,
			depth:      cb.depth + 1,
			data:       cb.data,
		}

		if _, ok := cb.builder.FileDeps[p]; ok {
			cb.builder.FileDeps[p][path] = struct{}{}
		} else {
			cb.builder.FileDeps[p] = map[string]struct{}{path: {}}
		}

		c, err := nestedCB.ProcessAsByte(p, fileData)
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
