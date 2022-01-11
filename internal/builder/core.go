package builder

import (
	"io/fs"
)

type Builder struct {
	RootFolder string
	ConfigFile string
	BuildDir   string
	SrcDir     string

	HTMLDirectory *string

	Config       *MainConf
	FileBuilders map[string]FileBuilder

	FileDeps map[string]map[string]struct{}
}

type FileBuilder interface {
	Init() error
	CanHandle(string, fs.FileInfo) bool
	Process(string, fs.FileInfo) error
}
