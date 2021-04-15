package main

type Config struct {
	LocalIP string `yaml:"local_ip"`
	Server  struct {
		Port string `yaml:"port"`
		Cert string `yaml:"cert"`
		Key  string `yaml:"key"`
	} `yaml:"server"`
}
