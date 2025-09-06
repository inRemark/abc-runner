package connection

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	
	"redis-runner/app/adapters/kafka/config"
)

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxConnections    int           // 最大连接数
	ProducerPoolSize  int           // 生产者池大小
	ConsumerPoolSize  int           // 消费者池大小
	ConnectionTimeout time.Duration // 连接超时
}

// ConnectionPool 连接池管理器
type ConnectionPool struct {
	config        *config.KafkaAdapterConfig
	poolConfig    PoolConfig
	
	// 生产者池
	producerPool chan *kafka.Writer
	producers    []*kafka.Writer
	
	// 消费者池
	consumerPool chan *kafka.Reader
	consumers    []*kafka.Reader
	
	// 管理客户端
	adminConn *kafka.Conn
	
	// 同步控制
	mutex   sync.RWMutex
	closed  bool
}

// NewConnectionPool 创建连接池
func NewConnectionPool(kafkaConfig *config.KafkaAdapterConfig, poolConfig PoolConfig) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		config:       kafkaConfig,
		poolConfig:   poolConfig,
		producerPool: make(chan *kafka.Writer, poolConfig.ProducerPoolSize),
		consumerPool: make(chan *kafka.Reader, poolConfig.ConsumerPoolSize),
		producers:    make([]*kafka.Writer, 0, poolConfig.ProducerPoolSize),
		consumers:    make([]*kafka.Reader, 0, poolConfig.ConsumerPoolSize),
	}
	
	// 初始化连接池
	if err := pool.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}
	
	return pool, nil
}

// initialize 初始化连接池
func (p *ConnectionPool) initialize() error {
	// 创建TLS配置
	var tlsConfig *tls.Config
	if p.config.Security.TLS.Enabled {
		var err error
		tlsConfig, err = p.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
	}
	
	// 创建SASL机制
	var saslMechanism sasl.Mechanism
	if p.config.Security.SASL.Enabled {
		var err error
		saslMechanism, err = p.createSASLMechanism()
		if err != nil {
			return fmt.Errorf("failed to create SASL mechanism: %w", err)
		}
	}
	
	// 初始化生产者池
	if err := p.initializeProducers(tlsConfig, saslMechanism); err != nil {
		return fmt.Errorf("failed to initialize producers: %w", err)
	}
	
	// 初始化消费者池
	if err := p.initializeConsumers(tlsConfig, saslMechanism); err != nil {
		return fmt.Errorf("failed to initialize consumers: %w", err)
	}
	
	// 初始化管理连接
	if err := p.initializeAdminConnection(tlsConfig, saslMechanism); err != nil {
		return fmt.Errorf("failed to initialize admin connection: %w", err)
	}
	
	return nil
}

// initializeProducers 初始化生产者池
func (p *ConnectionPool) initializeProducers(tlsConfig *tls.Config, saslMechanism sasl.Mechanism) error {
	for i := 0; i < p.poolConfig.ProducerPoolSize; i++ {
		writer := &kafka.Writer{
			Addr:                   kafka.TCP(p.config.Brokers...),
			Topic:                  "", // Topic will be set per message
			Balancer:               p.createBalancer(),
			MaxAttempts:            p.config.Producer.Retries + 1,
			BatchSize:              p.config.Producer.BatchSize,
			BatchTimeout:           p.config.Producer.LingerMs,
			ReadTimeout:            p.config.Producer.ReadTimeout,
			WriteTimeout:           p.config.Producer.WriteTimeout,
			RequiredAcks:           p.parseAcks(p.config.Producer.Acks),
			Async:                  false,
			Completion:             nil,
			Compression:            p.parseCompression(p.config.Producer.Compression),
			Logger:                 nil, // TODO: 集成日志系统
			ErrorLogger:           nil, // TODO: 集成日志系统
			Transport:             p.createTransport(tlsConfig, saslMechanism),
		}
		
		p.producers = append(p.producers, writer)
		p.producerPool <- writer
	}
	
	return nil
}

// initializeConsumers 初始化消费者池
func (p *ConnectionPool) initializeConsumers(tlsConfig *tls.Config, saslMechanism sasl.Mechanism) error {
	for i := 0; i < p.poolConfig.ConsumerPoolSize; i++ {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:                p.config.Brokers,
			Topic:                  p.config.Benchmark.DefaultTopic,
			GroupID:                p.config.Consumer.GroupID,
			MinBytes:               p.config.Consumer.FetchMinBytes,
			MaxBytes:               p.config.Consumer.FetchMaxBytes,
			MaxWait:                p.config.Consumer.FetchMaxWait,
			ReadBatchTimeout:       p.config.Consumer.ReadTimeout,
			StartOffset:            p.parseStartOffset(p.config.Consumer.InitialOffset),
			RebalanceTimeout:       p.config.Consumer.SessionTimeout,
			HeartbeatInterval:      p.config.Consumer.HeartbeatInterval,
			CommitInterval:         p.getCommitInterval(),
			PartitionWatchInterval: 1 * time.Second,
			WatchPartitionChanges:  true,
			Logger:                 nil, // TODO: 集成日志系统
			ErrorLogger:           nil, // TODO: 集成日志系统
			Dialer:                 p.createDialer(tlsConfig, saslMechanism),
		})
		
		p.consumers = append(p.consumers, reader)
		p.consumerPool <- reader
	}
	
	return nil
}

// initializeAdminConnection 初始化管理连接
func (p *ConnectionPool) initializeAdminConnection(tlsConfig *tls.Config, saslMechanism sasl.Mechanism) error {
	dialer := p.createDialer(tlsConfig, saslMechanism)
	
	ctx, cancel := context.WithTimeout(context.Background(), p.poolConfig.ConnectionTimeout)
	defer cancel()
	
	var err error
	p.adminConn, err = dialer.DialContext(ctx, "tcp", p.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to create admin connection: %w", err)
	}
	
	return nil
}

// createTLSConfig 创建TLS配置
func (p *ConnectionPool) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !p.config.Security.TLS.VerifySSL,
		ServerName:         p.config.Security.TLS.ServerName,
	}
	
	// 加载客户端证书
	if p.config.Security.TLS.CertFile != "" && p.config.Security.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(p.config.Security.TLS.CertFile, p.config.Security.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	
	// 加载CA证书
	if p.config.Security.TLS.CaFile != "" {
		// TODO: 实现CA证书加载
	}
	
	return tlsConfig, nil
}

// createSASLMechanism 创建SASL机制
func (p *ConnectionPool) createSASLMechanism() (sasl.Mechanism, error) {
	switch p.config.Security.SASL.Mechanism {
	case "PLAIN":
		return plain.Mechanism{
			Username: p.config.Security.SASL.Username,
			Password: p.config.Security.SASL.Password,
		}, nil
		
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, p.config.Security.SASL.Username, p.config.Security.SASL.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-256 mechanism: %w", err)
		}
		return mechanism, nil
		
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, p.config.Security.SASL.Username, p.config.Security.SASL.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-512 mechanism: %w", err)
		}
		return mechanism, nil
		
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", p.config.Security.SASL.Mechanism)
	}
}

// createDialer 创建连接拨号器
func (p *ConnectionPool) createDialer(tlsConfig *tls.Config, saslMechanism sasl.Mechanism) *kafka.Dialer {
	dialer := &kafka.Dialer{
		Timeout:   p.poolConfig.ConnectionTimeout,
		DualStack: true,
	}
	
	if tlsConfig != nil {
		dialer.TLS = tlsConfig
	}
	
	if saslMechanism != nil {
		dialer.SASLMechanism = saslMechanism
	}
	
	return dialer
}

// createTransport 创建传输层
func (p *ConnectionPool) createTransport(tlsConfig *tls.Config, saslMechanism sasl.Mechanism) *kafka.Transport {
	transport := &kafka.Transport{
		Dial: p.createDialer(tlsConfig, saslMechanism).DialFunc,
	}
	
	return transport
}

// createBalancer 创建负载均衡器
func (p *ConnectionPool) createBalancer() kafka.Balancer {
	switch p.config.Benchmark.PartitionStrategy {
	case "round_robin":
		return &kafka.RoundRobin{}
	case "hash":
		return &kafka.Hash{}
	case "random":
		// kafka-go没有内置的随机平衡器，使用round robin代替
		return &kafka.RoundRobin{}
	default:
		return &kafka.LeastBytes{}
	}
}

// parseAcks 解析Acks配置
func (p *ConnectionPool) parseAcks(acks string) kafka.RequiredAcks {
	switch acks {
	case "0":
		return kafka.RequireNone
	case "1":
		return kafka.RequireOne
	case "all", "-1":
		return kafka.RequireAll
	default:
		return kafka.RequireOne
	}
}

// parseCompression 解析压缩配置
func (p *ConnectionPool) parseCompression(compression string) kafka.Compression {
	switch compression {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	default:
		return 0 // 使用数值0代表无压缩
	}
}

// parseStartOffset 解析起始偏移配置
func (p *ConnectionPool) parseStartOffset(offset string) int64 {
	switch offset {
	case "earliest":
		return kafka.FirstOffset
	case "latest":
		return kafka.LastOffset
	default:
		return kafka.LastOffset
	}
}

// getCommitInterval 获取提交间隔
func (p *ConnectionPool) getCommitInterval() time.Duration {
	if p.config.Consumer.EnableAutoCommit {
		return p.config.Consumer.AutoCommitInterval
	}
	return 0 // 禁用自动提交
}

// GetProducer 获取生产者
func (p *ConnectionPool) GetProducer() (*kafka.Writer, error) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mutex.RUnlock()
	
	select {
	case producer := <-p.producerPool:
		return producer, nil
	case <-time.After(p.poolConfig.ConnectionTimeout):
		return nil, fmt.Errorf("timeout waiting for producer from pool")
	}
}

// ReturnProducer 归还生产者
func (p *ConnectionPool) ReturnProducer(producer *kafka.Writer) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		return
	}
	p.mutex.RUnlock()
	
	select {
	case p.producerPool <- producer:
		// 成功归还
	default:
		// 池已满，丢弃连接
	}
}

// GetConsumer 获取消费者
func (p *ConnectionPool) GetConsumer() (*kafka.Reader, error) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mutex.RUnlock()
	
	select {
	case consumer := <-p.consumerPool:
		return consumer, nil
	case <-time.After(p.poolConfig.ConnectionTimeout):
		return nil, fmt.Errorf("timeout waiting for consumer from pool")
	}
}

// ReturnConsumer 归还消费者
func (p *ConnectionPool) ReturnConsumer(consumer *kafka.Reader) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		return
	}
	p.mutex.RUnlock()
	
	select {
	case p.consumerPool <- consumer:
		// 成功归还
	default:
		// 池已满，丢弃连接
	}
}

// GetAdminConnection 获取管理连接
func (p *ConnectionPool) GetAdminConnection() *kafka.Conn {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if p.closed {
		return nil
	}
	
	return p.adminConn
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.closed {
		return nil
	}
	
	p.closed = true
	
	// 关闭所有生产者
	close(p.producerPool)
	for _, producer := range p.producers {
		if err := producer.Close(); err != nil {
			// 记录错误但继续关闭其他连接
		}
	}
	
	// 关闭所有消费者
	close(p.consumerPool)
	for _, consumer := range p.consumers {
		if err := consumer.Close(); err != nil {
			// 记录错误但继续关闭其他连接
		}
	}
	
	// 关闭管理连接
	if p.adminConn != nil {
		if err := p.adminConn.Close(); err != nil {
			// 记录错误
		}
	}
	
	return nil
}

// Stats 获取连接池统计信息
func (p *ConnectionPool) Stats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return map[string]interface{}{
		"producer_pool_size":     len(p.producerPool),
		"producer_pool_capacity": cap(p.producerPool),
		"consumer_pool_size":     len(p.consumerPool),
		"consumer_pool_capacity": cap(p.consumerPool),
		"total_producers":        len(p.producers),
		"total_consumers":        len(p.consumers),
		"closed":                 p.closed,
	}
}