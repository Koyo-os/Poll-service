package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Reqs struct {
		CreatePollRequestType string `yaml:"create_req_type"`
		UpdatePollRequestType string `yaml:"update_req_type"`
	} `yaml:"reqs"`
	Topic struct {
		Request  string `yaml:"req_topic"`
		Producer string `yaml:"producer_topic"`
	} `yaml:"topics"`
	GroupID  string `yaml:"group_id_name"`
	KafkaUrl string `yaml:"kafka_url"`
	Dsn      string `yaml:"dsn"`
}

func Init(path string) (*Config, error) {
	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error open file: %v", err)
	}

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return &cfg, nil
}
