package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Users    map[string]string
	Address  string
	Port     string
	Pages    map[string]string
	ArticlesPerPage int
	Protocol string
	Certfile string
	Keyfile  string
}

func LoadConfig(filename string) (*Config, error) {
	cont, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := Config{}
	err = yaml.Unmarshal(cont, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
