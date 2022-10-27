package builder

import (
	"io/fs"
)

type Builder struct {
	opts *BuilderOpts

	initialized bool

	rootFolder string
	buildDir   string
	srcDir     string

	currentLanguage string

	htmlDirectory *string
	varsDirectory *string

	fileBuilders      map[string]FileBuilder
	fileBuildersArray []FileBuilder

	fileDeps map[string]map[string]struct{}

	isSubBuilder bool
	subBuilders  map[string]*Builder // Used in multi lang scenarios
}

type BuilderOpts struct {
}

func NewBuilder(srcDir, buildDir, rootFolder string) *Builder {
	return &Builder{
		srcDir:     srcDir,
		buildDir:   buildDir,
		rootFolder: rootFolder,
	}
}

func (b *Builder) BuildDir() string {
	return b.buildDir
}

type BuildEnv struct {
	Lang           string
	AvaliableLangs []string
}

type FileBuilder interface {
	Init() error
	CanHandle(string, fs.FileInfo) bool
	Process(string, fs.FileInfo) error
}
