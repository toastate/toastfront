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
}

func (fb *FolderBuilder) Init() error {
	return nil
}

func (fb *FolderBuilder) CanHandle(path string, file fs.FileInfo) bool {
	return file.IsDir()
}

func (fb *FolderBuilder) Process(path string, file fs.FileInfo) error {
	fplist := strings.Split(path, string(filepath.Separator))
	for _, v := range fplist {
		if len(v) == 0 {
			continue
		}
		if v[0] == '.' || v[0] == '_' {
			return nil
		}
		if v == "includes" {
			return nil
		}
	}

	if *fb.builder.HTMLInSubFolder {
		if path == "html" {
			return nil
		}
		if strings.HasPrefix(path, "html/") {
			path = path[5:]
		}
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
