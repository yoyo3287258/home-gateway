package kafka

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/yoyo3287258/home-gateway/internal/config"
	"github.com/yoyo3287258/home-gateway/internal/model"
)

// Producer Kafka生产者
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer 创建Kafka生产者
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka生产者失败: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.RequestTopic,
	}, nil
}

// SendRequest 发送请求消息到Kafka
func (p *Producer) SendRequest(req *model.KafkaRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(req.TraceID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送Kafka消息失败: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// Consumer Kafka消费者
type Consumer struct {
	consumer      sarama.Consumer
	topic         string
	timeout       time.Duration
	pendingMu     sync.RWMutex
	pendingResps  map[string]chan *model.KafkaResponse
}

// NewConsumer 创建Kafka消费者
func NewConsumer(cfg *config.KafkaConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	c := &Consumer{
		consumer:     consumer,
		topic:        cfg.ResponseTopic,
		timeout:      cfg.ResponseTimeout,
		pendingResps: make(map[string]chan *model.KafkaResponse),
	}

	// 启动消费者协程
	go c.consumeLoop()

	return c, nil
}

// consumeLoop 消费响应消息的循环
func (c *Consumer) consumeLoop() {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		fmt.Printf("获取分区失败: %v\n", err)
		return
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("订阅分区 %d 失败: %v\n", partition, err)
			continue
		}

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				c.handleMessage(msg)
			}
		}(pc)
	}
}

// handleMessage 处理接收到的消息
func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	var resp model.KafkaResponse
	if err := json.Unmarshal(msg.Value, &resp); err != nil {
		fmt.Printf("解析响应消息失败: %v\n", err)
		return
	}

	c.pendingMu.RLock()
	ch, ok := c.pendingResps[resp.TraceID]
	c.pendingMu.RUnlock()

	if ok {
		// 非阻塞发送，防止超时后channel被关闭
		select {
		case ch <- &resp:
		default:
		}
	}
}

// WaitForResponse 等待指定TraceID的响应
func (c *Consumer) WaitForResponse(traceID string) (*model.KafkaResponse, error) {
	ch := make(chan *model.KafkaResponse, 1)

	c.pendingMu.Lock()
	c.pendingResps[traceID] = ch
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pendingResps, traceID)
		c.pendingMu.Unlock()
	}()

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(c.timeout):
		return nil, fmt.Errorf("等待响应超时（%v）", c.timeout)
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// Client Kafka客户端（封装生产者和消费者）
type Client struct {
	Producer *Producer
	Consumer *Consumer
}

// NewClient 创建Kafka客户端
func NewClient(cfg *config.KafkaConfig) (*Client, error) {
	producer, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	consumer, err := NewConsumer(cfg)
	if err != nil {
		producer.Close()
		return nil, err
	}

	return &Client{
		Producer: producer,
		Consumer: consumer,
	}, nil
}

// SendAndWait 发送请求并等待响应
func (c *Client) SendAndWait(req *model.KafkaRequest) (*model.KafkaResponse, error) {
	// 发送请求
	if err := c.Producer.SendRequest(req); err != nil {
		return nil, err
	}

	// 等待响应
	return c.Consumer.WaitForResponse(req.TraceID)
}

// Close 关闭客户端
func (c *Client) Close() error {
	var errs []error
	if err := c.Producer.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Consumer.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("关闭Kafka客户端失败: %v", errs)
	}
	return nil
}
