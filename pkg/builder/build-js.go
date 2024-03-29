package builder

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/toastate/toastfront/internal/helpers"
	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

var JSBuilderImportRegexp = regexp.MustCompile(`(?m)^import "local:\/\/(.*)";$`)
var JSBuilderImportHTMLVarsFuncRegexp = regexp.MustCompile(`(?m)toastfront\.pagevars\((.+)\)`)
var JSBuilderImportVarsFuncRegexp = regexp.MustCompile(`(?m)toastfront\.jsvars\(\)`)

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

	if jsData, ok := config.Config.BuilderConfig["javascript"]; ok {
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

	if cb.depth == 0 {
		cb.data = map[string]interface{}{}

		varsPath := filepath.Join(cb.builder.srcDir, cb.folder, cb.VarsFile)
		vf, err := os.Open(varsPath)
		if err != nil {
			if !os.IsNotExist(err) {
				tlogger.Warn("builder", "js", "msg", "Can't open js vars file", "file", varsPath, "err", err)
			}
		} else {
			defer vf.Close()
			err = json.NewDecoder(vf).Decode(&cb.data)
			if err != nil {
				tlogger.Error("builder", "js", "msg", "Can't decode js vars file", "file", varsPath, "err", err)
				return err
			}
		}
	}

	f, err := cb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	of, err := os.OpenFile(filepath.Join(cb.builder.buildDir, path), os.O_CREATE|os.O_RDWR, 0644)
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
		return nil, ErrTooDeep
	}
	f, err := os.ReadFile(filepath.Join(cb.builder.srcDir, path))
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

		fileData, err := os.Stat(filepath.Join(cb.builder.srcDir, p))
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

		if _, ok := cb.builder.fileDeps[p]; ok {
			cb.builder.fileDeps[p][path] = struct{}{}
		} else {
			cb.builder.fileDeps[p] = map[string]struct{}{path: {}}
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

	f = JSBuilderImportHTMLVarsFuncRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		p := string(JSBuilderImportHTMLVarsFuncRegexp.FindSubmatch(match)[1])

		p = strings.Trim(p, "\"")
		p = strings.Trim(p, "'")

		p = strings.ReplaceAll(p, "/", string(os.PathSeparator))

		if p[0] == os.PathSeparator {
			p = p[1:]
		}

		htmlBuilder := cb.builder.fileBuilders["html"].(*HTMLBuilder)
		pathData := htmlBuilder.GetPathDataDir(p)
		jsm, _ := helpers.MarshalJson(pathData)
		return bytes.TrimRight(jsm, "\n ")

		// fileData, err := os.Stat(filepath.Join(cb.builder.SrcDir, p))
		// if err != nil {
		// 	tlogger.Error("builder", "js", "msg", "vars error import", "sourcefile", path, "expectedfile", p, "err", err)
		// 	return []byte{'\n'}
		// }

		// return c
	})

	f = JSBuilderImportVarsFuncRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		env := os.Environ()
		for i := 0; i < len(env); i++ {
			spl := strings.Split(env[i], "=")
			if len(spl) == 2 {
				cb.data[spl[0]] = spl[1]
			}
		}

		jsm, _ := helpers.MarshalJson(cb.data)
		return bytes.TrimRight(jsm, "\n ")
	})

	return f, nil
}
