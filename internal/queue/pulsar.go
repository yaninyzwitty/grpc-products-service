package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/apache/pulsar-client-go/pulsar"
)

// PulsarMethods defines the interface for Pulsar-related operations
type PulsarMethods interface {
	CreatePulsarConnection(ctx context.Context) (pulsar.Client, error)
	CreatePulsarProducer(ctx context.Context, client pulsar.Client) (pulsar.Producer, error)
	CreatePulsarConsumer(ctx context.Context, client pulsar.Client, consumerTopic string) (pulsar.Consumer, error)
}

// PulsarConfig holds the configuration for the Pulsar connection
type PulsarConfig struct {
	URI       string
	Token     string
	TopicName string
}

// NewPulsar initializes and returns a PulsarConfig instance that implements PulsarMethods
func NewPulsar(cfg *PulsarConfig) PulsarMethods {
	return &PulsarConfig{
		URI:       cfg.URI,
		Token:     cfg.Token,
		TopicName: cfg.TopicName,
	}
}

// CreatePulsarConnection establishes a connection to the Pulsar server
func (c *PulsarConfig) CreatePulsarConnection(ctx context.Context) (pulsar.Client, error) {
	clientOptions := pulsar.ClientOptions{
		URL:            c.URI,
		Authentication: pulsar.NewAuthenticationToken(c.Token),
	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Pulsar client: %w", err)
	}

	slog.Info("Pulsar connection created successfully")

	return client, nil
}

// CreatePulsarProducer creates a new producer for a specified topic
func (c *PulsarConfig) CreatePulsarProducer(ctx context.Context, client pulsar.Client) (pulsar.Producer, error) {
	producerOptions := pulsar.ProducerOptions{
		Topic: c.TopicName,
	}

	producer, err := client.CreateProducer(producerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar producer: %w", err)
	}

	slog.Info("Pulsar producer created successfully", "topic", c.TopicName)

	return producer, nil
}

func (c *PulsarConfig) CreatePulsarConsumer(ctx context.Context, client pulsar.Client, consumerTopic string) (pulsar.Consumer, error) {
	consumerOptions := pulsar.ConsumerOptions{
		Topic:                       consumerTopic,
		SubscriptionName:            "my-subscription",
		SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	}

	consumer, err := client.Subscribe(consumerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar consumer: %w", err)
	}

	slog.Info("Pulsar consumer created successfully", "topic", c.TopicName)

	return consumer, nil
}
