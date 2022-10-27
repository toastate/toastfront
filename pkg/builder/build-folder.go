package builder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

type FolderBuilder struct {
	builder *Builder

	htmlVarsFolder string
}

func (fb *FolderBuilder) Init() error {
	tlogger.Debug("builder", "folder", "msg", "init")

	fb.htmlVarsFolder = "vars"

	if htmlData, ok := config.Config.BuilderConfig["html"]; ok {
		if data, ok := htmlData["vars_folder"]; ok {
			fb.htmlVarsFolder = data
		}
	}
	return nil
}

func (fb *FolderBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return file.IsDir()
}

func (fb *FolderBuilder) RewritePath(path string) string {
	if fb.builder.htmlDirectory != nil {
		if path == *fb.builder.htmlDirectory {
			return ""
		}
		if path == *fb.builder.varsDirectory {
			return ""
		}
		if strings.HasPrefix(path, *fb.builder.varsDirectory+string(os.PathSeparator)) {
			return ""
		}
		if strings.HasPrefix(path, *fb.builder.htmlDirectory+string(os.PathSeparator)) {
			path = path[len(*fb.builder.htmlDirectory)+1:]
		}
	}

	if strings.HasPrefix(path, fb.htmlVarsFolder+string(os.PathSeparator)) {
		return ""
	}

	return path
}

func (fb *FolderBuilder) Process(path string, file fs.FileInfo) error {
	path = fb.RewritePath(path)
	if path == "" {
		return nil
	}

	newFolder := filepath.Join(fb.builder.buildDir, path)

	err := os.MkdirAll(newFolder, 0755)
	if err != nil {
		tlogger.Error("builder", "folder", "file", path, "msg", "Failed to create folder", "err", err)
		return err
	}

	tlogger.Debug("builder", "folder", "file", newFolder, "msg", "Folder created")
	return nil
}
