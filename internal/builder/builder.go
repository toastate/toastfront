package builder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
)

func (b *Builder) Init() error {
	if b.RootFolder == "" {
		b.RootFolder = "."
	}
	if b.ConfigFile == "" {
		b.ConfigFile = "toastfront.json"
	}
	err := b.ReadConfig()
	if err != nil {
		return err
	}

	if b.CurrentLanguage == "" {
		b.CurrentLanguage = b.Config.RootLanguage
	}

	if b.BuildDir == "" {
		b.BuildDir = filepath.Join(b.RootFolder, b.Config.BuildDir)
	}
	if b.SrcDir == "" {
		b.SrcDir = filepath.Join(b.RootFolder, b.Config.SrcDir)
	}

	if b.FileDeps == nil {
		b.FileDeps = make(map[string]map[string]struct{})
	}

	if _, err := os.Stat(b.SrcDir); os.IsNotExist(err) {
		tlogger.Error("msg", "Src folder not found", "path", b.SrcDir, "err", err)
		return errors.New("Src folder not found")
	}

	if b.HTMLDirectory == nil {
		if b.Config.HTMLDir != "." {
			if b.Config.HTMLDir == "" { // Auto detect
				if f, _ := os.Stat(filepath.Join(b.SrcDir, "html")); f != nil && f.IsDir() {
					a := "html"
					b.HTMLDirectory = &a
				}
			} else { // Use config value
				b.HTMLDirectory = &b.Config.HTMLDir
			}
		} else {
			a := ""
			b.HTMLDirectory = &a
		}
	}

	if b.VarsDirectory == nil {
		if b.Config.VarsDir == "" {
			a := "html/vars"
			b.VarsDirectory = &a
		} else { // Use config value
			b.VarsDirectory = &b.Config.VarsDir
		}
	}

	b.FileBuilders = map[string]FileBuilder{
		"folder": &FolderBuilder{builder: b},
		"css":    &CSSBuilder{builder: b},
		"html":   &HTMLBuilder{builder: b},
		"js":     &JSBuilder{builder: b},
		// "vendor": &VendorBuilder{builder: b},
		"copy": &CopyBuilder{builder: b},
	}
	b.FileBuildersArray = []FileBuilder{
		b.FileBuilders["folder"],
		b.FileBuilders["css"],
		b.FileBuilders["html"],
		b.FileBuilders["js"],
		b.FileBuilders["copy"],
	}

	if len(b.SubBuilders) > 0 {
		for _, v := range b.SubBuilders {
			v.Init()
		}
	} else if !b.IsSubBuilder && len(b.Config.Languages) > 1 {
		buildDir := b.BuildDir
		b.SubBuilders = map[string]*Builder{}
		i := 0
		if b.Config.RootLanguage == "" {
			b.CurrentLanguage = b.Config.Languages[0]
			b.BuildDir = filepath.Join(buildDir, b.CurrentLanguage)
			i++
		}
		for ; i < len(b.Config.Languages); i++ {
			lg := b.Config.Languages[i]
			if lg == b.CurrentLanguage {
				continue
			}
			subBuilder := &Builder{
				RootFolder:      b.RootFolder,
				ConfigFile:      b.ConfigFile,
				HTMLDirectory:   b.HTMLDirectory,
				VarsDirectory:   b.VarsDirectory,
				Config:          b.Config,
				CurrentLanguage: lg,
				SrcDir:          b.SrcDir,
				BuildDir:        filepath.Join(buildDir, lg),
				IsSubBuilder:    true,
			}
			err := subBuilder.Init()
			if err != nil {
				return err
			}
			b.SubBuilders[lg] = subBuilder
		}
	}

	for _, v := range b.FileBuildersArray {
		err := v.Init()
		if err != nil {
			return err
		}
	}

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

func (b *Builder) Build() error {
	err := b.Init()
	if err != nil {
		return err
	}

	err = os.RemoveAll(b.BuildDir)
	if err != nil {
		tlogger.Error("msg", "Failed to remove build folder", "path", b.BuildDir, "err", err)
		return err
	}
	err = os.MkdirAll(b.BuildDir, 0755)
	if err != nil {
		tlogger.Error("msg", "Failed to create build folder", "path", b.BuildDir, "err", err)
		return err
	}

	tlogger.Info("msg", "Building started", "path", b.SrcDir)
	defer tlogger.Info("msg", "Building finished", "path", b.SrcDir)

	err = filepath.Walk(b.SrcDir, func(absolutepath string, info fs.FileInfo, err error) error {

		path, err := filepath.Rel(b.SrcDir, absolutepath)
		if err != nil {
			tlogger.Error("msg", "Failed to get relative path", "path", path, "err", err)
			return err
		}

		if b.ShouldHandle(path) {
			for _, v := range b.FileBuildersArray {
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

		for _, subBuilder := range b.SubBuilders {
			for _, v := range subBuilder.FileBuildersArray {
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
		for k, v := range b.FileBuilders {
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
