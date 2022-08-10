package builder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
)

type FolderBuilder struct {
	builder *Builder

	htmlVarsFolder string
}

func (fb *FolderBuilder) Init() error {
	tlogger.Debug("builder", "folder", "msg", "init")

	fb.htmlVarsFolder = "vars"

	if htmlData, ok := fb.builder.Config.BuilderConfig["html"]; ok {
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
	if fb.builder.HTMLDirectory != nil {
		if path == *fb.builder.HTMLDirectory {
			return ""
		}
		if path == *fb.builder.VarsDirectory {
			return ""
		}
		if strings.HasPrefix(path, *fb.builder.VarsDirectory+string(os.PathSeparator)) {
			return ""
		}
		if strings.HasPrefix(path, *fb.builder.HTMLDirectory+string(os.PathSeparator)) {
			path = path[len(*fb.builder.HTMLDirectory)+1:]
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

	newFolder := filepath.Join(fb.builder.BuildDir, path)

	err := os.MkdirAll(newFolder, 0755)
	if err != nil {
		tlogger.Error("builder", "folder", "file", path, "msg", "Failed to create folder", "err", err)
		return err
	}

	tlogger.Debug("builder", "folder", "file", newFolder, "msg", "Folder created")
	return nil
}
