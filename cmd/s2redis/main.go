package main

import (
	"s2redis"
)

func main() {
	db, err := s2redis.NewSingleStore(s2redis.DatabaseConfig{
		Host:     "172.17.0.4",
		Port:     "3306",
		Username: "root",
		Password: "test",
		Database: "kv",
	})
	if err != nil {
		panic(err)
	}

	server := s2redis.NewServer(db)

	if err := server.ListenAndServe("6379"); err != nil {
		panic(err)
	}
}
