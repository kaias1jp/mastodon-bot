package config

import (
	toml "github.com/BurntSushi/toml"
)

// DBConfig represents database connection configuration information.
type AppConfig struct { // toml内の名前を入れる
	Server       string `toml:"server"`
	ClientID     string `toml:"clientID"`
	ClientSecret string `toml:"clientSecret"`
	Email        string `toml:"email"`
	Password     string `toml:"password"`
}

// Config represents application configuration.
type Config struct { // toml内の名前を入れる
	APP AppConfig `toml:"mastodon"`
}

// NewConfig return configuration struct.
func NewConfig(path string, appMode string) (Config, error) {
	var conf Config

	confPath := path + appMode + ".toml" // tomlファイルを読み設定情報を取得
	if _, err := toml.DecodeFile(confPath, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
