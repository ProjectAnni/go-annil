package config

import (
	"github.com/go-yaml/yaml"
	uuid "github.com/satori/go.uuid"
	"os"
)

const configPath = "config.yml"

type Config struct {
	Secret   string `yaml:"secret"`
	Listen   string `yaml:"listen"`
	RepoRoot string `yaml:"root"`
}

var Cfg = Config{
	Secret:   uuid.NewV4().String(),
	Listen:   "0.0.0.0:8000",
	RepoRoot: "repo",
}

func Init() (err error) {
	if err := Load(); err != nil {
		err = Save()
	}
	_ = Save()
	return
}

func Save() error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	return encoder.Encode(Cfg)
}

func Load() error {
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	return decoder.Decode(&Cfg)
}
