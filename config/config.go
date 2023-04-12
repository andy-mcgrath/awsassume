package config

import (
	"fmt"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// LoadConfig loads accounts and roles configuration from ~/.aws/awsassume.toml
func LoadConfig() (*koanf.Koanf, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("~/.aws/awsassume.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return k, nil
}
