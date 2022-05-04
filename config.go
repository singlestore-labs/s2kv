package s2kv

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func LoadTOMLFiles(out interface{}, filenames []string) error {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Printf("toml file `%s` not found, skipping", filename)
			continue
		}
		_, err := toml.DecodeFile(filename, out)
		if err != nil {
			return err
		}
	}
	return nil
}
