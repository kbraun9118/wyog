package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Config struct {
	file *ini.File
}

func ReadConfig() (*Config, error) {
	xdgConfigHome := "~/.config"
	if configHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		xdgConfigHome = configHome
	}

	configs := make([]any, 0)
	if _, err := os.Stat("~/.gitconfig"); err == nil {
		configs = append(configs, "~/.gitconfig")
	}

	config, err := ini.Load(
		os.ExpandEnv(filepath.Join(xdgConfigHome, "git/config")),
		configs...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config files")
	}

	return &Config{config}, err
}

func (c *Config) User() string {
	if user, err := c.file.GetSection("user"); err == nil {
		name, nErr := user.GetKey("name")
		email, eErr := user.GetKey("email")
		if nErr == nil && eErr == nil {
			return fmt.Sprintf("%s <%s>", name, email)
		}
	}

	return ""
}
