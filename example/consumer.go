package main

type Handler struct{}

func (h *Handler) Handle(event interface{}) error {
	return nil
}

func main() {
	eventbus, err := coreos.NewAmqpClient(":5672")
	if err != nil {
		panic(err)
	}

	eventbus.Subscribe("TransactionFinished", new(Handler))

	// Block, since Subscibe() runs in its own goroutine
	select {}
}
