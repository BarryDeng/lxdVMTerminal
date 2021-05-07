package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

func main() {
	ConfigFile, err := os.Open("config.yml")
	if err != nil {
		log.Fatalf("open config.yml: %v", err)
	}
	defer ConfigFile.Close()

	var cfg Config
	decoder := yaml.NewDecoder(ConfigFile)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalf("decode config.yml: %v", err)
	}

	s := InitServer(&cfg)
	s.StartServer()
}
