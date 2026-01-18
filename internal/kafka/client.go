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

// Producer Kafka鐢熶骇鑰?
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer 鍒涘缓Kafka鐢熶骇鑰?
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("鍒涘缓Kafka鐢熶骇鑰呭け璐? %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.RequestTopic,
	}, nil
}

// SendRequest 鍙戦€佽姹傛秷鎭埌Kafka
func (p *Producer) SendRequest(req *model.KafkaRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("搴忓垪鍖栬姹傚け璐? %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(req.TraceID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("鍙戦€並afka娑堟伅澶辫触: %w", err)
	}

	return nil
}

// Close 鍏抽棴鐢熶骇鑰?
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// Consumer Kafka娑堣垂鑰?
type Consumer struct {
	consumer      sarama.Consumer
	topic         string
	timeout       time.Duration
	pendingMu     sync.RWMutex
	pendingResps  map[string]chan *model.KafkaResponse
}

// NewConsumer 鍒涘缓Kafka娑堣垂鑰?
func NewConsumer(cfg *config.KafkaConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("鍒涘缓Kafka娑堣垂鑰呭け璐? %w", err)
	}

	c := &Consumer{
		consumer:     consumer,
		topic:        cfg.ResponseTopic,
		timeout:      cfg.ResponseTimeout,
		pendingResps: make(map[string]chan *model.KafkaResponse),
	}

	// 鍚姩娑堣垂鑰呭崗绋?
	go c.consumeLoop()

	return c, nil
}

// consumeLoop 娑堣垂鍝嶅簲娑堟伅鐨勫惊鐜?
func (c *Consumer) consumeLoop() {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		fmt.Printf("鑾峰彇鍒嗗尯澶辫触: %v\n", err)
		return
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("璁㈤槄鍒嗗尯 %d 澶辫触: %v\n", partition, err)
			continue
		}

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				c.handleMessage(msg)
			}
		}(pc)
	}
}

// handleMessage 澶勭悊鎺ユ敹鍒扮殑娑堟伅
func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	var resp model.KafkaResponse
	if err := json.Unmarshal(msg.Value, &resp); err != nil {
		fmt.Printf("瑙ｆ瀽鍝嶅簲娑堟伅澶辫触: %v\n", err)
		return
	}

	c.pendingMu.RLock()
	ch, ok := c.pendingResps[resp.TraceID]
	c.pendingMu.RUnlock()

	if ok {
		// 闈為樆濉炲彂閫侊紝闃叉瓒呮椂鍚巆hannel琚叧闂?
		select {
		case ch <- &resp:
		default:
		}
	}
}

// WaitForResponse 绛夊緟鎸囧畾TraceID鐨勫搷搴?
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
		return nil, fmt.Errorf("绛夊緟鍝嶅簲瓒呮椂锛?v锛?, c.timeout)
	}
}

// Close 鍏抽棴娑堣垂鑰?
func (c *Consumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// Client Kafka瀹㈡埛绔紙灏佽鐢熶骇鑰呭拰娑堣垂鑰咃級
type Client struct {
	Producer *Producer
	Consumer *Consumer
}

// NewClient 鍒涘缓Kafka瀹㈡埛绔?
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

// SendAndWait 鍙戦€佽姹傚苟绛夊緟鍝嶅簲
func (c *Client) SendAndWait(req *model.KafkaRequest) (*model.KafkaResponse, error) {
	// 鍙戦€佽姹?
	if err := c.Producer.SendRequest(req); err != nil {
		return nil, err
	}

	// 绛夊緟鍝嶅簲
	return c.Consumer.WaitForResponse(req.TraceID)
}

// Close 鍏抽棴瀹㈡埛绔?
func (c *Client) Close() error {
	var errs []error
	if err := c.Producer.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Consumer.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("鍏抽棴Kafka瀹㈡埛绔け璐? %v", errs)
	}
	return nil
}
