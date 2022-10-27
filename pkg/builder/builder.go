package builder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

// Init is idempotent, multiple calls will only initialize the builder once
func (b *Builder) Init() error {
	if b.initialized {
		return nil
	}

	if b.rootFolder == "" {
		b.rootFolder = "."
	}

	if b.currentLanguage == "" {
		b.currentLanguage = config.Config.RootLanguage
	}

	if b.buildDir == "" {
		b.buildDir = filepath.Join(b.rootFolder, config.Config.BuildDir)
	}
	if b.srcDir == "" {
		b.srcDir = filepath.Join(b.rootFolder, config.Config.SrcDir)
	}

	if b.fileDeps == nil {
		b.fileDeps = make(map[string]map[string]struct{})
	}

	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		tlogger.Error("msg", "Src folder not found", "path", b.srcDir, "err", err)
		return errors.New("src folder not found")
	}

	if b.htmlDirectory == nil {
		if config.Config.HTMLDir != "." {
			if config.Config.HTMLDir == "" { // Auto detect
				if f, _ := os.Stat(filepath.Join(b.srcDir, "html")); f != nil && f.IsDir() {
					a := "html"
					b.htmlDirectory = &a
				}
			} else { // Use config value
				b.htmlDirectory = &config.Config.HTMLDir
			}
		} else {
			a := ""
			b.htmlDirectory = &a
		}
	}

	if b.varsDirectory == nil {
		if config.Config.VarsDir == "" {
			a := "html/vars"
			b.varsDirectory = &a
		} else { // Use config value
			b.varsDirectory = &config.Config.VarsDir
		}
	}

	b.fileBuilders = map[string]FileBuilder{
		"folder": &FolderBuilder{builder: b},
		"css":    &CSSBuilder{builder: b},
		"html":   &HTMLBuilder{builder: b},
		"js":     &JSBuilder{builder: b},
		// "vendor": &VendorBuilder{builder: b},
		"copy": &CopyBuilder{builder: b},
	}
	b.fileBuildersArray = []FileBuilder{
		b.fileBuilders["folder"],
		b.fileBuilders["css"],
		b.fileBuilders["html"],
		b.fileBuilders["js"],
		b.fileBuilders["copy"],
	}

	if len(b.subBuilders) > 0 {
		for _, v := range b.subBuilders {
			v.Init()
		}
	} else if !b.isSubBuilder && len(config.Config.Languages) > 1 {
		buildDir := b.buildDir
		b.subBuilders = map[string]*Builder{}
		i := 0
		if config.Config.RootLanguage == "" {
			b.currentLanguage = config.Config.Languages[0]
			b.buildDir = filepath.Join(buildDir, b.currentLanguage)
			i++
		}
		for ; i < len(config.Config.Languages); i++ {
			lg := config.Config.Languages[i]
			if lg == b.currentLanguage {
				continue
			}
			subBuilder := &Builder{
				rootFolder:      b.rootFolder,
				htmlDirectory:   b.htmlDirectory,
				varsDirectory:   b.varsDirectory,
				currentLanguage: lg,
				srcDir:          b.srcDir,
				buildDir:        filepath.Join(buildDir, lg),
				isSubBuilder:    true,
			}
			err := subBuilder.Init()
			if err != nil {
				return err
			}
			b.subBuilders[lg] = subBuilder
		}
	}

	for _, v := range b.fileBuildersArray {
		err := v.Init()
		if err != nil {
			return err
		}
	}

	b.initialized = true
	return nil
}

func (b *Builder) ShouldHandle(name string) bool {
	folderList := strings.Split(name, string(filepath.Separator))
	for _, v := range folderList {
		if v == "includes" {
			return false
		}
		if len(v) > 0 && (v[0] == '.' || v[0] == '_') {
			return false
		}
	}
	return true
}

func (b *Builder) Build(opts ...*BuilderOpts) error {
	if len(opts) > 0 {
		b.opts = opts[0]
	}

	err := os.RemoveAll(b.buildDir)
	if err != nil {
		<-time.After(time.Millisecond * 20)
		err = os.RemoveAll(b.buildDir)
		if err != nil {
			<-time.After(time.Millisecond * 20)
			err = os.RemoveAll(b.buildDir)
			tlogger.Error("msg", "Failed to remove build folder", "path", b.buildDir, "err", err)
		}
	}

	err = os.MkdirAll(b.buildDir, 0755)
	if err != nil {
		tlogger.Error("msg", "Failed to create build folder", "path", b.buildDir, "err", err)
		return err
	}

	tlogger.Info("msg", "Building started", "path", b.srcDir)
	defer tlogger.Info("msg", "Building finished", "path", b.srcDir)

	err = filepath.Walk(b.srcDir, func(absolutepath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		path, err := filepath.Rel(b.srcDir, absolutepath)
		if err != nil {
			tlogger.Error("msg", "Failed to get relative path", "path", path, "err", err)
			return err
		}

		if b.ShouldHandle(path) {
			for _, v := range b.fileBuildersArray {
				if v.CanHandle(path, info) {
					err = v.Process(path, info)
					if err != nil {
						tlogger.Error("msg", "Error processing file", "path", path, "error", err)
						return err
					}

					break
				}
			}
		}

		for _, subBuilder := range b.subBuilders {
			for _, v := range subBuilder.fileBuildersArray {
				if v.CanHandle(path, info) {
					err = v.Process(path, info)
					if err != nil {
						tlogger.Error("msg", "Error processing file", "path", path, "error", err)
						return err
					}

					break
				}
			}
		}
		return nil
	})

	return err
}

func (b *Builder) BuildSingle(path string, info fs.FileInfo) error {
	if b.ShouldHandle(path) {
		for k, v := range b.fileBuilders {
			if v.CanHandle(path, info) {
				err := v.Process(path, info)
				if err != nil {
					tlogger.Error("msg", "Error processing file", "builder", k, "path", path, "error", err)
					return err
				}
			}
		}
	}
	return nil
}
