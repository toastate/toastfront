package builder

import (
	"encoding/json"
	htemplate "html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	ttemplate "text/template"

	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

var CSSBuilderImportRegexp = regexp.MustCompile(`(?m)^\s*@import "local:\/\/(.*)";\s*$`)

type CSSBuilder struct {
	builder *Builder
	depth   int // To avoid infinite recursive imports
	data    map[string]interface{}

	extension string
	folder    string
	varsFile  string
}

func (cb *CSSBuilder) Init() error {
	tlogger.Debug("builder", "css", "msg", "init")

	cb.varsFile = "vars.json"
	cb.folder = "css"
	cb.extension = ".css"

	if cssData, ok := config.Config.BuilderConfig["css"]; ok {
		if data, ok := cssData["vars_file"]; ok {
			cb.varsFile = data
		}
		if data, ok := cssData["folder"]; ok {
			cb.folder = data
		}
		if data, ok := cssData["ext"]; ok {
			cb.extension = data
		}
	}

	varsPath := filepath.Join(cb.builder.srcDir, cb.folder, cb.varsFile)
	f, err := os.Open(varsPath)
	if err != nil {
		tlogger.Warn("builder", "css", "msg", "Can't open css vars file", "file", varsPath, "err", err)
	} else {
		defer f.Close()
		err = json.NewDecoder(f).Decode(&cb.data)
		if err != nil {
			tlogger.Error("builder", "css", "msg", "Can't decode css vars file", "file", varsPath, "err", err)
			return err
		}
	}

	return nil
}

func (cb *CSSBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return cb.IsCssFile(path, file)
}

func (cb *CSSBuilder) IsCssFile(path string, file fs.FileInfo) bool {
	pathSplit := strings.Split(path, string(filepath.Separator))
	if pathSplit[0] != cb.folder {
		return false
	}
	return filepath.Ext(file.Name()) == cb.extension
}

func (cb *CSSBuilder) Process(path string, file fs.FileInfo) error {
	tlogger.Debug("builder", "css", "msg", "processing", "file", path)

	f, err := cb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	of, err := os.OpenFile(filepath.Join(cb.builder.buildDir, path), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		tlogger.Error("builder", "css", "msg", "output file creation", "file", path, "err", err)
		return err
	}
	defer of.Close()

	// wr := filewriter.Writer("text/css", of)

	env := os.Environ()
	for i := 0; i < len(env); i++ {
		spl := strings.Split(env[i], "=")
		if len(spl) == 2 {
			cb.data[spl[0]] = spl[1]
		}
	}

	if config.Config.UnsafeVars {
		t, err := ttemplate.New(path).Delims(`"{{`, `}}"`).Parse(string(f))
		if err != nil {
			tlogger.Error("builder", "css", "msg", "temple", "file", path, "err", err)
			return err
		}

		err = t.Execute(of, cb.data)
		if err != nil {
			tlogger.Error("builder", "css", "msg", "templater", "file", path, "err", err)
			return err
		}
	} else {
		t, err := htemplate.New(path).Delims(`"{{`, `}}"`).Parse(string(f))
		if err != nil {
			tlogger.Error("builder", "css", "msg", "temple", "file", path, "err", err)
			return err
		}

		err = t.Execute(of, cb.data)
		if err != nil {
			tlogger.Error("builder", "css", "msg", "templater", "file", path, "err", err)
			return err
		}
	}

	return nil
}

func (cb *CSSBuilder) ProcessAsByte(path string, file fs.FileInfo) ([]byte, error) {
	if cb.depth > 5 {
		tlogger.Debug("builder", "css", "msg", "file error", "file", path, "err", "reached max recursion depth of 5, import loop ?")
		return nil, ErrTooDeep
	}
	f, err := os.ReadFile(filepath.Join(cb.builder.srcDir, path))
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

		p = filepath.Join(cb.folder, p)

		fileData, err := os.Stat(filepath.Join(cb.builder.srcDir, p))
		if err != nil {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}

		if !cb.IsCssFile(p, fileData) {
			tlogger.Error("builder", "css", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", "import types mismatched")
			return []byte{'\n'}
		}

		nestedCB := &CSSBuilder{
			folder:    cb.folder,
			extension: cb.extension,
			varsFile:  cb.varsFile,
			builder:   cb.builder,
			depth:     cb.depth + 1,
			data:      cb.data,
		}

		if _, ok := cb.builder.fileDeps[p]; ok {
			cb.builder.fileDeps[p][path] = struct{}{}
		} else {
			cb.builder.fileDeps[p] = map[string]struct{}{path: {}}
		}

		c, err := nestedCB.ProcessAsByte(p, fileData)
		if err != nil {
			tlogger.Error("builder", "css", "msg", "file error process", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}
		if len(c) == 0 || c[len(c)-1] != '\n' {
			c = append(c, '\n')
		}

		return c
	})

	return f, nil
}
