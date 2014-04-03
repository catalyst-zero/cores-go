package client

import (
	"fmt"
	amqp "github.com/streadway/amqp"
	"math/rand"
	"os"
	"time"
)

func init() {
	// NOTE: We may need to seed more properly
	rand.Seed(time.Now().UTC().UnixNano())
}

func getQueueName(eventName, consumerGroup string) string {
	return fmt.Sprintf("%s.%s", consumerGroup, eventName)
}

func getConsumerId(consumerGroup, eventName string) string {
	var hostname string

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	pid := os.Getpid()

	return fmt.Sprintf("cores-go#%s.%s#%d@%s/%d", consumerGroup, eventName, pid, hostname, rand.Uint32())
}

func newAmqpSubscription(eventName, consumerGroup string, autoAck bool) *amqpSubscriber {
	queueName := getQueueName(eventName, consumerGroup)

	return &amqpSubscriber{
		ConsumerGroup: consumerGroup,
		QueueName:     queueName,
		EventName:     eventName,
		AutoAck:       autoAck,

		EventChan: make(chan Event),
	}
}

type amqpSubscriber struct {
	QueueName     string
	ConsumerGroup string
	EventName     string
	AutoAck       bool

	EventChan chan Event

	consumerId string
	channel    *amqp.Channel
}

func (c *amqpSubscriber) init(publishExchange string, connection *amqp.Connection) error {
	assertNotNil(connection)

	// Generate a new consumerId each time, we connect - otherwise we run into errors when acknowledging messages
	consumerId := getConsumerId(c.ConsumerGroup, c.EventName)
	c.consumerId = consumerId

	channel, err := connection.Channel()
	if err != nil {
		return err
	}

	// Error listener
	errorChan := channel.NotifyClose(make(chan *amqp.Error))
	go func() {
		// NOTE: We need this in case the channel gets broken (but not the connection), e.g.
		// when an invalid message is acknowledged.
		for err := range errorChan {
			fmt.Printf("ERROR %s: In queue-channel '%s': %v\n", consumerId, c.QueueName, err)
			if cerr := c.init(publishExchange, connection); cerr != nil {
				fmt.Printf("ERROR: Failed to reestablish channel for queue %s: %v\n", c.QueueName, cerr)
			}
		}
	}()

	// Step 1) Queue
	if _, err := channel.QueueDeclare(c.QueueName, true, false, false, false, nil); err != nil {
		return err
	}

	// Step 2) QueueBinding
	if err := channel.QueueBind(c.QueueName, c.EventName, publishExchange, false, nil); err != nil {
		return err
	}

	deliveryChan, err := channel.Consume(c.QueueName, consumerId, c.AutoAck, false, false, false, nil)
	if err != nil {
		return err
	}

	c.channel = channel
	go c.runEventTransformer(deliveryChan)
	return nil
}

func (c *amqpSubscriber) runEventTransformer(deliveryChan <-chan amqp.Delivery) {
	for delivery := range deliveryChan {
		event := &amqpEvent{delivery}
		c.EventChan <- event
	}
}

func (c *amqpSubscriber) Events() <-chan Event {
	return c.EventChan
}

func (c *amqpSubscriber) Close() error {
	defer close(c.EventChan)

	// NOTE: This should close the deliveryChannel, which quits the loop in Run(), which stops this subscriber
	if err := c.channel.Cancel(c.consumerId, false); err != nil {
		return err
	}
	return nil
}
