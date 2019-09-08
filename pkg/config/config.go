package config

import (
	"errors"
	"io"

	"github.com/naoina/toml"
)

type Config struct {
	Server struct {
		Development       bool   `toml:"development"`
		ListenAddress     string `toml:"listen_address"`
		SigningKey        string `toml:"signing_key"`
		MapboxAccessToken string `toml:"mapbox_access_token"`
	} `toml:"server"`
	DB struct {
		Driver string `toml:"driver"`
		DSN    string `toml:"dsn"`
	} `toml:"db"`
}

func Read(r io.Reader) (*Config, error) {
	config := new(Config)
	if err := toml.NewDecoder(r).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func Check(config *Config) error {
	if config.Server.SigningKey == "" {
		return errors.New("missing server.signing_key")
	}
	return nil
}
