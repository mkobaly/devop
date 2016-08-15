package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type UserCredential struct {
	URL      string
	Username string
	Password string
}

type Config struct {
	Jira    UserCredential
	Octopus struct {
		URL       string
		Webapikey string
	}
	Teamcity UserCredential
}

//NewConfig creates a new Configuration object needed
func NewConfig(configPath string) *Config {
	//config := Config{}
	var config = new(Config)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err.Error())
	}
	return config
}
