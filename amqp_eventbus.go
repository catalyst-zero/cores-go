package client

import (
	"encoding/json"
	"fmt"
	amqp "github.com/streadway/amqp"
	"math/rand"
	"time"
)

const (
	CONTENT_TYPE_JSON = "application/json"
)

type amqpConnectionDial func() (*amqp.Connection, error)

func NewAmqpEventBus(url string) (EventBus, error) {
	factory := func() (*amqp.Connection, error) {
		return amqp.Dial(url)
	}
	return newAmqpEventBus(factory)
}

func NewAmqpEventBusConfig(url string, config amqp.Config) (EventBus, error) {
	factory := func() (*amqp.Connection, error) {
		return amqp.DialConfig(url, config)
	}
	return newAmqpEventBus(factory)
}

func newAmqpEventBus(factory amqpConnectionDial) (EventBus, error) {
	client := &amqpClient{
		amqpPublisher{
			PublishExchange: "hutch",
		},
		amqpSubscriptionManager{
			PublishExchange: "hutch",
			Subscribers:     make([]*amqpSubscriber, 0),
		},
	}
	runAmqpInitializer(factory, client)

	return client, nil
}

func runAmqpInitializer(factory amqpConnectionDial, client *amqpClient) {
	connection, err := factory()
	if err != nil {
		fmt.Printf("AMQP Error: Failed to dial: %v\n", err)
		go retryLater(factory, client)
		return
	}

	closeChan := connection.NotifyClose(make(chan *amqp.Error))

	if err := client.Init(connection); err != nil {
		fmt.Printf("AMQP Error: Failed to init amqpEventBus: %v\n", err)
		go retryLater(factory, client)
		return
	}

	go func() {
		for err := range closeChan {
			fmt.Printf("AMQP Error: %v\n", err) // TODO: Cleanup?
			go runAmqpInitializer(factory, client)
		}
	}()
}

func retryLater(factory amqpConnectionDial, client *amqpClient) {
	time.Sleep(5 * time.Second)
	runAmqpInitializer(factory, client)
}

type amqpClient struct {
	amqpPublisher
	amqpSubscriptionManager
}

func (bus *amqpClient) Init(connection *amqp.Connection) error {
	if err := bus.amqpPublisher.Init(connection); err != nil {
		return err
	}
	if err := bus.amqpSubscriptionManager.Init(connection); err != nil {
		return err
	}
	return nil
}

func (bus *amqpClient) Close() error {
	if err := bus.amqpPublisher.Close(); err != nil {
		return err
	}
	return bus.amqpSubscriptionManager.Close()
}

type amqpSubscriptionManager struct {
	PublishExchange string
	connection      *amqp.Connection
	Subscribers     []*amqpSubscriber
}

func (sub *amqpSubscriptionManager) Init(connection *amqp.Connection) error {
	sub.connection = connection

	if sub.Subscribers != nil && len(sub.Subscribers) > 0 {
		for _, subscriber := range sub.Subscribers {
			subscriber.init(sub.PublishExchange, connection)
		}
	}
	return nil
}

// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
func (sub *amqpSubscriptionManager) Subscribe(eventName, consumerGroup string, autoAck bool) (Subscription, error) {
	queueName := sub.queueName(eventName, consumerGroup)
	consumerTag := fmt.Sprintf("cores-go#%s.%s#%d", consumerGroup, eventName, rand.Uint32())

	subscription := newAmqpSubscription(queueName, consumerTag, eventName, autoAck)
	if err := subscription.init(sub.PublishExchange, sub.connection); err != nil {
		return nil, err
	}
	sub.Subscribers = append(sub.Subscribers, subscription)
	return subscription, nil
}

func (client *amqpSubscriptionManager) queueName(eventName, consumerGroup string) string {
	return fmt.Sprintf("%s.%s", consumerGroup, eventName)
}

// Stops all subscribes and closes all internal connections
func (client *amqpSubscriptionManager) Close() error {
	for _, sub := range client.Subscribers {
		sub.Close()
	}
	return client.connection.Close()
}

// AmqpPublisher contains all the logic needed to send events to the broker.
type amqpPublisher struct {
	PublishExchange string
	publishChannel  *amqp.Channel
}

func (client *amqpPublisher) Init(connection *amqp.Connection) error {
	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	client.publishChannel = channel

	return client.ensurePublishingExchange()
}

// Publish sends the given payload to all consumers subscribed to the given event.
func (client *amqpPublisher) Publish(eventName string, payload interface{}) error {
	if client.publishChannel == nil {
		return fmt.Errorf("connection currently not available.")
	}

	exchangeName := client.PublishExchange
	routingKey := eventName

	data, err := json.Marshal(payload)
	if err != nil {
		// Failed to encode payload
		return err
	}

	publishing := amqp.Publishing{
		ContentType:     CONTENT_TYPE_JSON,
		ContentEncoding: "UTF-8",
		DeliveryMode:    2,
		Body:            data,
	}

	return client.publishChannel.Publish(exchangeName, routingKey, false, false, publishing)
}

func (p *amqpPublisher) ensurePublishingExchange() error {
	return p.publishChannel.ExchangeDeclare(p.PublishExchange, "topic", true, false, false, false, nil)
}

func (p *amqpPublisher) Close() error {
	if p.publishChannel != nil {
		return p.publishChannel.Close()
	}
	return nil
}
