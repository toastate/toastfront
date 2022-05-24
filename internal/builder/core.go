package builder

import (
	"io/fs"
)

type Builder struct {
	RootFolder string
	ConfigFile string
	BuildDir   string
	SrcDir     string

	CurrentLanguage string

	HTMLDirectory *string
	VarsDirectory *string

	Config            *MainConf
	FileBuilders      map[string]FileBuilder
	FileBuildersArray []FileBuilder

	FileDeps map[string]map[string]struct{}

	SubBuilders map[string]*Builder // Used in multi lang scenarios
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
