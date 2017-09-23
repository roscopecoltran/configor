package main

import (
	// "github.com/k0kubun/pp"
	"github.com/roscopecoltran/configor"
)

var Config = struct {
	APPName string `default:"app name"`

	DB struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Port     int    `default:"3306"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}

	Oauth2 struct {
		Github struct {
			Token        string `json:"personal_token" yaml:"personal_token"`
			ClientKey    string `json:"client_id" yaml:"client_id"`
			ClientSecret string `json:"client_secret" yaml:"client_secret"`
		} `json:"github" yaml:"github"`
	} `json:"oauth2" yaml:"oauth2"`
}{}

func main() {
	configor.Load(&Config, "config.yml")
	// pp.Print(Config)
	configor.Dump(Config, "all", "yaml,toml,json", "./dump")

}
