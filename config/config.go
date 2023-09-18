package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path       string   `yaml:"path"`
	SSHCommand string   `yaml:"ssh_command"`
	Debug      bool     `yaml:"debug"`
	Defaults   Defaults `yaml:"defaults"`
}

type Defaults struct {
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	SSHPass    string `yaml:"ssh_password"`
	BecomePass string `yaml:"become_password"`
	PrivateKey string `yaml:"private_key"`
}

// Read config from file system
func Read(configPath string) (*Config, error) {
	configb, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configb, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
