package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

const fileName = "config.yaml"

// Configuration structure
type Config struct {
	GitLab     CfgService     `yaml:"GitLab"`
	Confluence CfgService     `yaml:"Confluence"`
	Projects   map[int]string `yaml:"Projects"`
}

type CfgService struct {
	Endpoint string `yaml:"Endpoint"`
	User     string `yaml:"User"`
	Pass     string `yaml:"Pass"`
	Space    string `yaml:"Space"`
	Page     int    `yaml:"Page"`
}

// Generate config
func New() *Config {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return &c
}
