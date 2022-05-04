package s2kv_test

import (
	"flag"
	"s2kv"
	"testing"
)

var flagConfigPath = flag.String("config", "config.example.toml", "path to an optional config file")

func GetSingleStore(t *testing.T) *s2kv.SingleStore {
	configPath := *flagConfigPath
	config := s2kv.Config{}
	if configPath != "" {
		err := s2kv.LoadTOMLFiles(&config, []string{configPath})
		if err != nil {
			t.Fatal(err)
		}
	}

	db, err := s2kv.NewSingleStore(config.Database)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
