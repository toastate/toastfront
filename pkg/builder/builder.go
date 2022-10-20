package builder

import "github.com/toastate/toastfront/internal/builder"

type Builder interface {
	Init() error
	Build() error
}

func NewBuilder(srcDir string, buildDir string, rootFolder string) Builder {
	return builder.NewBuilder(srcDir, buildDir, rootFolder)
}
