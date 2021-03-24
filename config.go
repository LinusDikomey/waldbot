package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GuildId string			`yaml:"guildId"`
	StatsChannelId string	`yaml:"statsChannelId"`
	CertFile string			`yaml:"certFile"`
	KeyFile string			`yaml:"keyFile"`
}

var config Config = Config {}

func loadConfig() {
	file, err := os.Open("./data/config.yaml")
	defer file.Close()
	if err != nil {
		log.Fatal("Could not open config.yaml:", err)
	}
	bytes, _ := ioutil.ReadAll(file)
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal("Error while parsing config.yaml: ", err)
	}
}