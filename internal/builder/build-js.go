package builder

import (
	"encoding/json"
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
)

var JSBuilderImportRegexp = regexp.MustCompile(`(?m)^import "local:\/\/(.*)";$`)
var JSBuilderImportFuncRegexp = regexp.MustCompile(`(?m)toastfront\.pagevars\((.+)\)`)

type JSBuilder struct {
	builder *Builder
	depth   int // To avoid infinite recursive imports
	data    map[string]interface{}

	extension string
	folder    string
	VarsFile  string
}

func (cb *JSBuilder) Init() error {
	tlogger.Debug("builder", "js", "msg", "init")

	cb.VarsFile = "vars.json"
	cb.folder = "js"
	cb.extension = ".js"

	if jsData, ok := cb.builder.Config.BuilderConfig["js"]; ok {
		if data, ok := jsData["vars_file"]; ok {
			cb.VarsFile = data
		}
		if data, ok := jsData["folder"]; ok {
			cb.folder = data
		}
		if data, ok := jsData["ext"]; ok {
			cb.extension = data
		}
	}

	varsPath := filepath.Join(cb.builder.SrcDir, cb.folder, cb.VarsFile)
	f, err := os.Open(varsPath)
	if err != nil {
		tlogger.Warn("builder", "js", "msg", "Can't open js vars file", "file", varsPath, "err", err)
	} else {
		defer f.Close()
		err = json.NewDecoder(f).Decode(&cb.data)
		if err != nil {
			tlogger.Error("builder", "js", "msg", "Can't decode js vars file", "file", varsPath, "err", err)
			return err
		}
	}

	return nil
}

func (cb *JSBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return cb.IsJsFile(path, file)
}

func (cb *JSBuilder) IsJsFile(path string, file fs.FileInfo) bool {
	pathSplit := strings.Split(path, string(filepath.Separator))
	if pathSplit[0] != cb.folder {
		return false
	}
	return filepath.Ext(file.Name()) == cb.extension
}

func (cb *JSBuilder) Process(path string, file fs.FileInfo) error {
	tlogger.Debug("builder", "js", "msg", "processing", "file", path)

	f, err := cb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	of, err := os.OpenFile(filepath.Join(cb.builder.BuildDir, path), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		tlogger.Error("builder", "js", "msg", "output file creation", "file", path, "err", err)
		return err
	}
	defer of.Close()

	wr := of

	wr.Write(f)
	wr.Close()

	return nil
}

func (cb *JSBuilder) GetInternal(path string) []byte {
	return nil
}

func (cb *JSBuilder) ProcessAsByte(path string, file fs.FileInfo) ([]byte, error) {
	if cb.depth > 5 {
		tlogger.Debug("builder", "js", "msg", "file error", "file", path, "err", "reached max recursion depth of 5, import loop ?")
		return nil, errors.New("Too deep")
	}
	f, err := ioutil.ReadFile(filepath.Join(cb.builder.SrcDir, path))
	if err != nil {
		tlogger.Error("builder", "js", "msg", "file error", "file", path, "err", err)
		return nil, err
	}

	f = replaceWindowsCarriageReturn(f)

	f = JSBuilderImportRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		p := string(JSBuilderImportRegexp.FindSubmatch(match)[1])

		// Split on /, check if begins by __internal, if so, load html vars, json

		p = strings.ReplaceAll(p, "/", string(os.PathSeparator))

		if p[0] == os.PathSeparator {
			p = p[1:]
		}

		p = filepath.Join(cb.folder, p)

		fileData, err := os.Stat(filepath.Join(cb.builder.SrcDir, p))
		if err != nil {
			tlogger.Error("builder", "js", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}

		if !cb.IsJsFile(p, fileData) {
			tlogger.Error("builder", "js", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", "import types mismatched")
			return []byte{'\n'}
		}

		nestedCB := &JSBuilder{
			folder:    cb.folder,
			extension: cb.extension,
			VarsFile:  cb.VarsFile,
			builder:   cb.builder,
			depth:     cb.depth + 1,
			data:      cb.data,
		}

		if _, ok := cb.builder.FileDeps[p]; ok {
			cb.builder.FileDeps[p][path] = struct{}{}
		} else {
			cb.builder.FileDeps[p] = map[string]struct{}{path: {}}
		}

		c, err := nestedCB.ProcessAsByte(p, fileData)
		if err != nil {
			tlogger.Error("builder", "js", "msg", "file error process", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}
		if c[len(c)-1] != '\n' {
			c = append(c, '\n')
		}

		return c
	})

	f = JSBuilderImportFuncRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		p := string(JSBuilderImportFuncRegexp.FindSubmatch(match)[1])

		p = strings.Trim(p, "\"")
		p = strings.Trim(p, "'")

		p = strings.ReplaceAll(p, "/", string(os.PathSeparator))

		if p[0] == os.PathSeparator {
			p = p[1:]
		}

		htmlBuilder := cb.builder.FileBuilders["html"].(*HTMLBuilder)
		pathData := htmlBuilder.GetPathDataDir(p)
		jsm, _ := MarshalJson(pathData)
		return jsm

		// fileData, err := os.Stat(filepath.Join(cb.builder.SrcDir, p))
		// if err != nil {
		// 	tlogger.Error("builder", "js", "msg", "vars error import", "sourcefile", path, "expectedfile", p, "err", err)
		// 	return []byte{'\n'}
		// }

		// return c
	})

	return f, nil
}
