package client

import (
	"encoding/json"
	"fmt"
	amqp "github.com/streadway/amqp"
)

func NewAmqpClient(addr, consumergroup string) (Client, error) {
	connection, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	client := &amqpClient{
		consumerGroup:   consumergroup,
		connection:      connection,
		publishExchange: "hutch",
		publishChannel:  channel,
		subscribers:     make([]*amqpConsumer, 0),
	}

	if err := client.ensurePublishingExchange(); err != nil {
		return nil, err
	}

	return client, nil
}

type amqpClient struct {
	consumerGroup   string
	connection      *amqp.Connection
	publishChannel  *amqp.Channel
	publishExchange string
	subscribers     []*amqpConsumer
}

// Returns the name of this consumer group, e.g. "orlok" or "hades"
func (client *amqpClient) GetConsumerGroup() string {
	return client.consumerGroup
}

func (client *amqpClient) Publish(eventName string, payload interface{}) error {
	exchangeName := client.publishExchange
	routingKey := eventName

	data, err := json.Marshal(payload)
	if err != nil {
		// Failed to encode payload
		return err
	}

	publishing := amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "UTF-8",
		DeliveryMode:    2,
		Body:            data,
	}

	return client.publishChannel.Publish(exchangeName, routingKey, false, false, publishing)
}

func (client *amqpClient) ensurePublishingExchange() error {
	return channel.ExchangeDeclare(client.publishExchange, "topic", true, false, false, false, nil)
}

// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
func (client *amqpClient) Subscribe(eventName string, autoAck bool, handler Handler) error {
	queueName := client.queueName(eventName)

	channel, err := client.connection.Channel()
	if err != nil {
		return err
	}

	if err = client.ensureConsumer(channel, eventName); err != nil {
		return err
	}

	consumerTag := fmt.Sprintf("cores-go#%s.%s", client.consumerGroup, eventName)
	deliveryChan, err := channel.Consume(queueName, consumerTag, autoAck, false, false, false, nil)

	if err != nil {
		return err
	}

	subscriber := &amqpConsumer{
		deliveryChan: deliveryChan,
		closeChan:    make(chan bool),
		handler:      handler,
	}
	go subscriber.Run()
	client.subscribers = append(client.subscribers, subscriber)
	return nil
}

func (client *amqpClient) queueName(eventName string) string {
	return fmt.Sprintf("%s.%s", client.consumerGroup, eventName)
}

func (client *amqpClient) ensureConsumer(channel *amqp.Channel, eventName string) error {
	queueName := client.queueName(eventName)

	// Step 1) Queue
	if _, err := channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	// Step 2) QueueBinding
	if err := channel.QueueBind(queueName, "", client.publishExchange, false, nil); err != nil {
		return err
	}
	return nil
}

// Stops all subscribes and closes all internal connections
func (client *amqpClient) Close() error {
	for _, sub := range client.subscribers {
		sub.Close()
	}
	return client.connection.Close()
}

type amqpConsumer struct {
	deliveryChan <-chan amqp.Delivery
	closeChan    chan bool
	channel      *amqp.Channel
	handler      Handler
}

func (c *amqpConsumer) Run() {
	for {
		select {
		case <-c.closeChan:
			// TODO: Properly close channel?
			return
		case delivery := <-c.deliveryChan:
			fmt.Sprintf("%v\n", delivery)

			event := &amqpEvent{&delivery}
			c.handler.Handle(event)
		}
	}
}

func (c *amqpConsumer) Close() {
	c.closeChan <- true
}

type amqpEvent struct {
	delivery *amqp.Delivery
}

func (e *amqpEvent) GetCorrelationId() string {
	return e.delivery.CorrelationId
}
func (e *amqpEvent) Parse(value interface{}) error {
	return json.Unmarshal(e.delivery.Body, value)
}

func (e *amqpEvent) Ack() error {
	return e.delivery.Ack(false)
}
func (e *amqpEvent) Nack(requeue bool) error {
	return e.delivery.Nack(false, requeue)
}
func (e *amqpEvent) Reject(requeue bool) error {
	return e.delivery.Reject(requeue)
}
