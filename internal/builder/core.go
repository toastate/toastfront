package builder

import (
	"io/fs"
	"sync"
)

type Builder struct {
	FileDeps        map[string]map[string]struct{}
	RootFolder      string
	BuildDir        string
	SrcDir          string
	HTMLInSubFolder *bool
	Config          *MainConf
	Initer          sync.Once
	FileBuilders    map[string]FileBuilder
}

type FileBuilder interface {
	Init() error
	CanHandle(string, fs.FileInfo) bool
	Process(string, fs.FileInfo) error
}

type MainConf struct {
	// Languages      []string            `json:"languages,omitempty"`
	// LanguagesNames map[string][]string `json:"languages_names,omitempty"`
	// RootLanguage   string              `json:"root_language,omitempty"`
}
