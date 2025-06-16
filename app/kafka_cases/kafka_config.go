package kafkaCases

import (
	"gopkg.in/yaml.v2"
	"os"
)

type KafkaConfig struct {
	Brokers  []string            `yaml:"brokers"`
	Topic    string              `yaml:"topic"`
	Consumer KafkaConsumerConfig `yaml:"consumer"`
	Producer KafkaProducerConfig `yaml:"producer"`
	TLS      KafkaTLSConfig      `yaml:"tls"`
	SASL     KafkaSASLConfig     `yaml:"sasl"`
}

type KafkaConsumerConfig struct {
	GroupID       string `yaml:"group_id"`
	InitialOffset string `yaml:"initial_offset"` // earliest / latest
	MinBytes      int    `yaml:"min_bytes"`
	MaxBytes      int    `yaml:"max_bytes"`
	ReadTimeout   string `yaml:"read_timeout"`
	WriteTimeout  string `yaml:"write_timeout"`
}

type KafkaProducerConfig struct {
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	Acks         string `yaml:"acks"` // no_ack / wait_for_local / wait_for_all
	MaxAttempts  int    `yaml:"max_attempts"`
}

type KafkaTLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CaFile   string `yaml:"ca_file"`
}

type KafkaSASLConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Mechanism string `yaml:"mechanism"` // PLAIN / SCRAM-SHA-256 / SCRAM-SHA-512
}

func LoadKafkaConfigDefault() *KafkaConfig {
	cfg, err := LoadKafkaConfig("conf/kafka.yaml")
	if err != nil {
		panic(err)
	}
	return cfg
}

func LoadKafkaConfig(path string) (*KafkaConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg KafkaConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
