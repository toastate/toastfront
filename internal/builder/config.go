package builder

import (
	"encoding/json"
	"os"

	"github.com/toastate/toastfront/internal/tlogger"
)

type MainConf struct {
	BuildDir      string                       `json:"build_directory,omitempty"`
	SrcDir        string                       `json:"source_directory,omitempty"`
	HTMLDir       string                       `json:"html_directory,omitempty"`
	VarsDir       string                       `json:"vars_directory,omitempty"`
	RootLanguage  string                       `json:"root_language,omitempty"`
	Languages     []string                     `json:"languages,omitempty"`
	LanguageMode  string                       `json:"language_mode,omitempty"`
	BuilderConfig map[string]map[string]string `json:"builder_config,omitempty"`
}

var DefaultMainConf = &MainConf{
	BuildDir:     "build",
	SrcDir:       "src",
	RootLanguage: "en",
	Languages: []string{
		"en",
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

func (b *Builder) ReadConfig() error {
	if b.Config != nil {
		return nil
	}

	_, err := os.Stat(b.ConfigFile)

	{
		jsm, _ := json.Marshal(DefaultMainConf)
		json.Unmarshal(jsm, &b.Config)
	}

	if !os.IsNotExist(err) {
		if err != nil {
			tlogger.Error("msg", "Error reading config file", "err", err)
			return err
		}

		f, err := os.Open(b.ConfigFile)
		if err != nil {
			tlogger.Error("msg", "Error reading config file", "err", err)
			return err
		}
		defer f.Close()

		err = json.NewDecoder(f).Decode(b.Config)
		if err != nil {
			tlogger.Error("msg", "Error unmarshaling config file", "err", err)
			return err
		}

		return nil
	}
	return nil
}
