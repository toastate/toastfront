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
	if fb.builder.HTMLDirectory != nil {
		if path == *fb.builder.HTMLDirectory {
			return nil
		}
		if strings.HasPrefix(path, *fb.builder.HTMLDirectory+"/") {
			path = path[len(*fb.builder.HTMLDirectory)+1:]
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
