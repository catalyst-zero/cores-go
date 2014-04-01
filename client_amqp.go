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

	client := &amqpClient{
		consumerGroup: consumergroup,
		connection:    connection,
		subscribers:   make([]*amqpConsumer, 0),
	}
	return client, nil
}

type amqpClient struct {
	consumerGroup string
	connection    *amqp.Connection
	subscribers   []*amqpConsumer
}

// Returns the name of this consumer group, e.g. "orlok" or "hades"
func (client *amqpClient) GetConsumerGroup() string {
	return client.consumerGroup
}

// Publish an event to all consumer groups
func (client *amqpClient) CreateProducer(options ProducerOptions) (Producer, error) {
	channel, err := client.connection.Channel()
	if err != nil {
		return nil, err
	}

	if err = channel.ExchangeDeclare(options.EventName, "fanout", true, false, false, false, nil); err != nil {
		// Failed to create exchange (or ensure it already exists properly)
		return nil, err
	}

	return &amqpProducer{options, channel}, nil
}

// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
func (client *amqpClient) Subscribe(eventName string, autoAck bool, handler Handler) error {
	channel, err := client.connection.Channel()
	if err != nil {
		return err
	}

	if err = client.ensureExchange(channel, eventName); err != nil {
		return err
	}
	if err = client.ensureConsumer(channel, eventName); err != nil {
		return err
	}

	consumerTag := fmt.Sprintf("cores-go#%s.%s", client.consumerGroup, eventName)
	deliveryChan, err := channel.Consume(client.queueName(eventName), consumerTag, autoAck, false, false, false, nil)
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
func (client *amqpClient) ensureExchange(channel *amqp.Channel, eventName string) error {
	return channel.ExchangeDeclare(eventName, "fanout", true, false, false, false, nil)
}

func (client *amqpClient) ensureConsumer(channel *amqp.Channel, eventName string) error {
	queueName := client.queueName(eventName)

	// Step 1) Queue
	if _, err := channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	// Step 2) QueueBinding
	if err := channel.QueueBind(queueName, "", eventName, false, nil); err != nil {
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

type amqpProducer struct {
	Options ProducerOptions
	Channel *amqp.Channel
}

func (p *amqpProducer) Send(payload interface{}) error {
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

	return p.Channel.Publish(p.Options.EventName, "", false, false, publishing)
}
