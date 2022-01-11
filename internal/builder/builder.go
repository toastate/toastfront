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
		}
	}

	b.FileBuilders = map[string]FileBuilder{
		"folder": &FolderBuilder{builder: b},
		"css":    &CSSBuilder{builder: b},
	}

	for _, v := range b.FileBuilders {
		err := v.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) ShouldHandle(name string) bool {
	folderList := strings.Split(name, string(filepath.Separator))
	if folderList[0] == "vendor" {
		return false
	}
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
			for k, v := range b.FileBuilders {
				if v.CanHandle(path, info) {
					err = v.Process(path, info)
					if err != nil {
						tlogger.Error("msg", "Error processing file", "builder", k, "path", path, "error", err)
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
