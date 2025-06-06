package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Reqs struct {
		SetClosedRequestType  string `yaml:"set_closed_req_type"`
		CreateVoteRequestType string `yaml:"create_vote_req_type"`
		CreatePollRequestType string `yaml:"create_req_type"`
		UpdatePollRequestType string `yaml:"update_req_type"`
	} `yaml:"reqs"`
	RabbitmqUrl     string `yaml:"rabbitmq_url"`
	RequestExchange string `yaml:"req_exchange"`
	VoteExchange    string `yaml:"vote_exchange"`
	OutputExcange   string `yaml:"out_exchange"`
	Dsn             string `yaml:"dsn"`
	QueueName       string `yaml:"queue_name"`
}

func Init(path string) (*Config, error) {
	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error open file: %v", err)
	}

	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	cfg.QueueName = "que"

	return &cfg, nil
}
