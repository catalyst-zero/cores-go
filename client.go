package client

type Handler interface {
	Handle(event Event)
}

type Event interface {
	GetCorrelationId() string

	// Parse parses the message into the given object.
	Parse(object interface{}) error

	Ack() error
	Nack(requeue bool) error
	Reject(requeue bool) error
}

type Client interface {
	// Returns the name of this consumer group, e.g. "orlok" or "hades"
	GetConsumerGroup() string

	// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
	Subscribe(eventName string, autoAck bool, handler Handler) error

	// Stops all subscribes and closes all internal connections
	Close() error

	CreateProducer(options ProducerOptions) (Producer, error)
}

type ProducerOptions struct {
	EventName string
}

type Producer interface {
	Send(payload interface{}) error
}
