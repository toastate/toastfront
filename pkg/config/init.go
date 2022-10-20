package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Config = DefaultConfiguration

var DefaultConfiguration = &Configuration{
	UnsafeVars:   false,
	BuildDir:     "build",
	SrcDir:       "src",
	RootLanguage: "en",
	Languages: []string{
		"en",
	},
	ServeConfig: ServeConfiguration{
		Redirect404: "",
		Port:        8100,
	},
	LanguageMode: "unique", // Any of unique, subfolder, folder
	HTMLDir:      "html",
	VarsDir:      "html/vars",
	BuilderConfig: map[string]map[string]string{
		"css": {
			"ext":       ".css",
			"folder":    "css",
			"vars_file": "config.json",
		},
		"vendor": {
			"folder": "vendor",
		},
		"assets": {
			"folder": "assets",
		},
		"html": {
			"ext": ".html",
		},
		"javascript": {
			"folder": "js",
			"ext":    ".js",
		},
	},
}

type Configuration struct {
	UnsafeVars    bool                         `json:"unsafe_vars,omitempty"`
	BuildDir      string                       `json:"build_directory,omitempty"`
	SrcDir        string                       `json:"source_directory,omitempty"`
	HTMLDir       string                       `json:"html_directory,omitempty"`
	VarsDir       string                       `json:"vars_directory,omitempty"`
	RootLanguage  string                       `json:"root_language,omitempty"`
	Languages     []string                     `json:"languages,omitempty"`
	LanguageMode  string                       `json:"language_mode,omitempty"`
	BuilderConfig map[string]map[string]string `json:"builder_config,omitempty"`
	ServeConfig   ServeConfiguration           `json:"serve_config,omitempty"`
}

type ServeConfiguration struct {
	Redirect404 string `json:"redirect_404"`
	Port        int    `json:"port"`
}

func Init(configpath string) error {
	if configpath == "" {
		configpath = "toastfront.json"
	}

	_, err := os.Stat(configpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("could not access configuration file %s: %v", configpath, err)
		}

		return nil
	}

	f, err := os.Open(configpath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(Config)
	if err != nil {
		return err
	}

	return nil
}
