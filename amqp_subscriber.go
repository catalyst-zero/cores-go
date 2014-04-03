package client

import (
	amqp "github.com/streadway/amqp"
)

func newAmqpSubscription(queueName, consumerId, eventName string, autoAck bool) *amqpSubscriber {
	return &amqpSubscriber{
		QueueName:  queueName,
		ConsumerId: consumerId,
		EventName:  eventName,
		AutoAck:    autoAck,

		EventChan: make(chan Event),
	}
}

type amqpSubscriber struct {
	QueueName  string
	ConsumerId string
	EventName  string
	AutoAck    bool

	EventChan chan Event

	channel *amqp.Channel
}

func (c *amqpSubscriber) init(publishExchange string, connection *amqp.Connection) error {
	assertNotNil(connection)

	channel, err := connection.Channel()
	if err != nil {
		return err
	}

	// Step 1) Queue
	if _, err := channel.QueueDeclare(c.QueueName, true, false, false, false, nil); err != nil {
		return err
	}

	// Step 2) QueueBinding
	if err := channel.QueueBind(c.QueueName, c.EventName, publishExchange, false, nil); err != nil {
		return err
	}

	deliveryChan, err := channel.Consume(c.QueueName, c.ConsumerId, c.AutoAck, false, false, false, nil)
	if err != nil {
		return err
	}

	c.channel = channel
	go c.runEventTransformer(deliveryChan)
	return nil
}

func (c *amqpSubscriber) runEventTransformer(deliveryChan <-chan amqp.Delivery) {
	for delivery := range deliveryChan {
		event := &amqpEvent{&delivery}
		c.EventChan <- event
	}
}

func (c *amqpSubscriber) Events() <-chan Event {
	return c.EventChan
}

func (c *amqpSubscriber) Close() error {
	defer close(c.EventChan)

	// NOTE: This should close the deliveryChannel, which quits the loop in Run(), which stops this subscriber
	if err := c.channel.Cancel(c.ConsumerId, false); err != nil {
		return err
	}
	return nil
}
