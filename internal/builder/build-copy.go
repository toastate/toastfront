package builder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/toastate/toastfront/internal/tlogger"
)

type CopyBuilder struct {
	builder *Builder
}

func (cp *CopyBuilder) Init() error {
	tlogger.Debug("builder", "copy", "msg", "init")

	return nil
}

func (cp *CopyBuilder) CanHandle(path string, file fs.FileInfo) bool {
	if path == *cp.builder.VarsDirectory {
		return false
	}
	if strings.HasPrefix(path, *cp.builder.VarsDirectory+"/") {
		return false
	}

	return !file.IsDir() && cp.IsAssetsFile(path, file)
}

func (cp *CopyBuilder) IsAssetsFile(path string, file fs.FileInfo) bool {
	return true
}

func (cp *CopyBuilder) RewritePath(path string) string {
	return path
}

func (cp *CopyBuilder) Process(path string, file fs.FileInfo) error {
	os.MkdirAll(filepath.Join(cp.builder.BuildDir, filepath.Dir(path)), 0755)
	_, err := copyFile(filepath.Join(cp.builder.SrcDir, path), filepath.Join(cp.builder.BuildDir, path))
	return err
}
