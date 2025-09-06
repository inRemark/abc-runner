package test

import (
	"testing"
	
	"redis-runner/app/adapters/kafka/config"
)

func TestKafkaAdapterConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.KafkaAdapterConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.KafkaAdapterConfig{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test-client",
				Producer: config.ProducerConfig{
					Acks:        "all",
					Retries:     3,
					BatchSize:   16384,
					Compression: "snappy",
				},
				Consumer: config.ConsumerConfig{
					GroupID:         "test-group",
					AutoOffsetReset: "latest",
					FetchMinBytes:   1024,
					FetchMaxBytes:   52428800,
				},
				Performance: config.PerformanceConfig{
					ConnectionPoolSize: 10,
					ProducerPoolSize:   5,
					ConsumerPoolSize:   5,
				},
				Benchmark: config.BenchmarkConfig{
					DefaultTopic: "test-topic",
					Total:        1000,
					Parallels:    10,
					ReadPercent:  50,
				},
			},
			wantErr: false,
		},
		{
			name: "empty brokers",
			config: &config.KafkaAdapterConfig{
				Brokers: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("KafkaAdapterConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigLoader_LoadFromDefault(t *testing.T) {
	loader := config.NewConfigLoader()
	cfg := loader.LoadFromDefault()

	if cfg == nil {
		t.Fatal("LoadFromDefault() returned nil")
	}

	if len(cfg.Brokers) == 0 {
		t.Error("Default config should have brokers")
	}
}