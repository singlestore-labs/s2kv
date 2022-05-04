package main

import (
	"flag"
	"log"
	"s2kv"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.example.toml", "path to an optional config file")
	flag.Parse()

	config := s2kv.Config{}
	if configPath != "" {
		err := s2kv.LoadTOMLFiles(&config, []string{configPath})
		if err != nil {
			log.Fatal(err)
		}
	}

	db, err := s2kv.NewSingleStore(config.Database)
	if err != nil {
		panic(err)
	}

	server := s2kv.NewServer(db)

	if err := server.ListenAndServe("6379"); err != nil {
		log.Fatal(err)
	}
}
