package client

import (
	"encoding/json"
	"fmt"

	amqp "github.com/streadway/amqp"
)

type amqpEvent struct {
	delivery amqp.Delivery
}

func (e *amqpEvent) GetCorrelationId() string {
	return e.delivery.CorrelationId
}

func (e *amqpEvent) Parse(value interface{}) error {
	if e.delivery.ContentType != CONTENT_TYPE_JSON {
		return fmt.Errorf("Unexpected content-type: %s", e.delivery.ContentType)
	}
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
