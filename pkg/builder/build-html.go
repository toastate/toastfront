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

	"github.com/toastate/toastfront/internal/helpers"
	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

var HTMLBuilderImportRegexp = regexp.MustCompile(`(?m)<!--\s*#import\s+(.*)\s*-->`)

type HTMLBuilder struct {
	builder  *Builder
	depth    int // To avoid infinite recursive imports
	baseData map[string]interface{}

	extension  string
	folder     string
	varsFolder string
}

func (cb *HTMLBuilder) Init() error {
	tlogger.Debug("builder", "html", "msg", "init")

	cb.varsFolder = config.DefaultConfiguration.VarsDir
	cb.folder = config.DefaultConfiguration.HTMLDir
	cb.extension = config.DefaultConfiguration.BuilderConfig["html"]["ext"]

	cb.folder = *cb.builder.htmlDirectory

	if cb.builder.varsDirectory != nil {
		cb.varsFolder = *cb.builder.varsDirectory
		if cb.varsFolder == "." {
			cb.varsFolder = ""
		}
	}

	if htmlData, ok := config.Config.BuilderConfig["html"]; ok {

		if data, ok := htmlData["ext"]; ok {
			cb.extension = data
		}
	}

	return nil
}

func (cb *HTMLBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return cb.IsHtmlFile(path, file)
}

func (cb *HTMLBuilder) IsHtmlFile(path string, file fs.FileInfo) bool {
	pathSplit := strings.Split(path, string(filepath.Separator))
	if cb.folder != "" {
		if pathSplit[0] != cb.folder {
			return false
		}
	}
	return filepath.Ext(file.Name()) == cb.extension
}

func (cb *HTMLBuilder) RewritePath(path string) string {
	if cb.builder.htmlDirectory != nil {
		if strings.HasPrefix(path, *cb.builder.htmlDirectory+string(filepath.Separator)) {
			path = path[len(*cb.builder.htmlDirectory)+1:]
		}
	}
	return path
}

func (cb *HTMLBuilder) GetPathData(path string) map[string]interface{} {
	varsDir := path[:len(path)-len(cb.extension)]
	out := cb.GetPathDataDir(varsDir)

	env := os.Environ()
	for i := 0; i < len(env); i++ {
		spl := strings.Split(env[i], "=")
		if len(spl) == 2 {
			out[spl[0]] = spl[1]
		}
	}

	return out
}

func (cb *HTMLBuilder) GetPathDataDir(varsDir string) map[string]interface{} {
	out := make(map[string]interface{})
	{
		bt, _ := helpers.MarshalJson(cb.baseData)
		json.Unmarshal(bt, &out)
	}

	varsPath := ""
	varsPath = filepath.Join(cb.builder.srcDir, cb.varsFolder, varsDir)

	{
		varsFile := filepath.Join(varsPath, "common.json")

		f, err := os.Open(varsFile)
		if err == nil {
			secondaryOut := make(map[string]interface{})
			err := json.NewDecoder(f).Decode(&secondaryOut)
			if err != nil {
				tlogger.Error("builder", "html", "msg", "Can't decode html vars file", "file", varsFile, "err", err)
			}
			for k, v := range secondaryOut {
				out[k] = v
			}
			f.Close()
		}
	}

	{

		varsFile := filepath.Join(varsPath, "lang-"+cb.builder.currentLanguage+".json")
		f, err := os.Open(varsFile)
		if err == nil {
			secondaryOut := make(map[string]interface{})
			err := json.NewDecoder(f).Decode(&secondaryOut)
			if err != nil {
				tlogger.Error("builder", "html", "msg", "Can't decode html vars file", "file", varsFile, "err", err)
			}
			for k, v := range secondaryOut {
				out[k] = v
			}
			f.Close()
		}
	}

	return out
}

func (cb *HTMLBuilder) Process(path string, file fs.FileInfo) error {
	tlogger.Debug("builder", "html", "msg", "processing", "file", path)

	if cb.depth == 0 {
		cb.baseData = map[string]interface{}{}

		varsPath := filepath.Join(cb.builder.srcDir, cb.varsFolder)

		{
			varsFile := filepath.Join(varsPath, "common.json")
			f, err := os.Open(varsFile)
			if err == nil {
				err = json.NewDecoder(f).Decode(&cb.baseData)
				f.Close()
				if err != nil {
					tlogger.Error("builder", "html", "msg", "Can't decode html vars file", "file", varsFile, "err", err)
					return err
				}
			}
		}
		{
			varsFile := filepath.Join(varsPath, "lang-"+cb.builder.currentLanguage+".json")
			f, err := os.Open(varsFile)
			if err == nil {
				tmp := make(map[string]interface{})
				err = json.NewDecoder(f).Decode(&tmp)
				f.Close()
				if err != nil {
					tlogger.Error("builder", "html", "msg", "Can't decode html vars file", "file", varsFile, "err", err)
					return err
				}
				for k, v := range tmp {
					cb.baseData[k] = v
				}
			}
		}
	}

	f, err := cb.ProcessAsByte(path, file)
	if err != nil {
		return err
	}

	pathOut := cb.RewritePath(path)

	of, err := os.OpenFile(filepath.Join(cb.builder.buildDir, pathOut), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		tlogger.Error("builder", "html", "msg", "output file creation", "file", pathOut, "err", err)
		return err
	}
	defer of.Close()

	// wr := filewriter.Writer("text/html", of)
	pathData := cb.GetPathData(pathOut)

	wr := of

	if config.Config.UnsafeVars {
		t, err := ttemplate.New(path).Delims(`<!--#`, `-->`).Parse(string(f))

		if err != nil {
			tlogger.Error("builder", "html", "msg", "templater", "file", path, "err", err)
			return err
		}

		err = t.Execute(wr, pathData)
		if err != nil {
			tlogger.Error("builder", "html", "msg", "templater", "file", path, "err", err)
			return err
		}
	} else {
		t, err := htemplate.New(path).Delims(`<!--#`, `-->`).Parse(string(f))

		if err != nil {
			tlogger.Error("builder", "html", "msg", "templater", "file", path, "err", err)
			return err
		}

		err = t.Execute(wr, pathData)
		if err != nil {
			tlogger.Error("builder", "html", "msg", "templater", "file", path, "err", err)
			return err
		}

	}

	return nil
}

func (cb *HTMLBuilder) ProcessAsByte(path string, file fs.FileInfo) ([]byte, error) {

	if cb.depth > 5 {
		tlogger.Debug("builder", "html", "msg", "file error", "file", path, "err", "reached max recursion depth of 5, import loop ?")
		return nil, ErrTooDeep
	}
	f, err := os.ReadFile(filepath.Join(cb.builder.srcDir, path))
	if err != nil {
		tlogger.Error("builder", "html", "msg", "file error", "file", path, "err", err)
		return nil, err
	}

	f = replaceWindowsCarriageReturn(f)

	f = HTMLBuilderImportRegexp.ReplaceAllFunc(f, func(match []byte) []byte {
		p := string(HTMLBuilderImportRegexp.FindSubmatch(match)[1])

		p = strings.ReplaceAll(p, "/", string(os.PathSeparator))

		if p[0] == os.PathSeparator {
			p = p[1:]
		}

		if cb.folder != "" {
			p = filepath.Join(cb.folder, p)
		}

		fileData, err := os.Stat(filepath.Join(cb.builder.srcDir, p))
		if err != nil {
			tlogger.Error("builder", "html", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}

		if !cb.IsHtmlFile(p, fileData) {
			tlogger.Error("builder", "html", "msg", "file error import", "sourcefile", path, "expectedfile", p, "err", "import types mismatched")
			return []byte{'\n'}
		}

		nestedCB := &HTMLBuilder{
			folder:     cb.folder,
			extension:  cb.extension,
			varsFolder: cb.varsFolder,
			builder:    cb.builder,
			depth:      cb.depth + 1,
			baseData:   cb.baseData,
		}

		if _, ok := cb.builder.fileDeps[p]; ok {
			cb.builder.fileDeps[p][path] = struct{}{}
		} else {
			cb.builder.fileDeps[p] = map[string]struct{}{path: {}}
		}

		c, err := nestedCB.ProcessAsByte(p, fileData)
		if err != nil {
			tlogger.Error("builder", "html", "msg", "file error process", "sourcefile", path, "expectedfile", p, "err", err)
			return []byte{'\n'}
		}
		if c[len(c)-1] != '\n' {
			c = append(c, '\n')
		}

		return c
	})

	return f, nil
}
