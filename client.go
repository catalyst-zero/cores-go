package client

type Handler interface {
	Handle(event interface{}) error
}

type Client interface {
	// Returns the name of this consumer group, e.g. "orlok" or "hades"
	GetConsumerGroup() string

	// Publish an event to all consumer groups
	Broadcast(eventName string, payload interface{}) error

	// Start receiving events. Events are distributed in the current consumer-group, so not every consumer receives all events.
	Subscribe(eventName string, handler Handler) error

	// Stops all subscribes and closes all internal connections
	Close() error
}
