package builder

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetMainConfigIfAvailable(rootFolder string) (*MainConf, error) {
	configPath := filepath.Join(rootFolder, "toastfront.json")
	_, err := os.Stat(configPath)
	if err != os.ErrNotExist {
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		c := &MainConf{}

		err = json.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}

		return c, nil
	}
	return nil, nil
}

// func (c *MainConf) IsMultiLang() bool {
// 	if c.Languages != nil && len(c.Languages) > 0 {
// 		return true
// 	}
// 	return false
// }
