package main

import (
	"fmt"
	cores "github.com/catalyst-zero/cores-go"
	"time"
)

var _ = time.Now

type TransactionFinishedEvent struct {
	Id string `json:"id"`
}

type Handler struct {
	counter int
}

func (h *Handler) DoWork(transaction TransactionFinishedEvent) error {
	// fmt.Printf("> %s\n", transaction.Id)
	h.counter++

	if h.counter > 3 {
		h.counter = 0
		return fmt.Errorf("Something went wrong ...")
	} else {
		return nil
	}
}

func (h *Handler) Handle(event cores.Event) {
	// fmt.Printf("Event received: %v\n", event)

	var transaction TransactionFinishedEvent

	if err := event.Parse(&transaction); err != nil {
		if err := event.Reject(false); err != nil {
			fmt.Printf("ERROR: Reject failed: %v\n", err)
			return
		}
		fmt.Printf("ERROR: Event#Parse() failed: %v\n", err)
		return
	}

	if err := h.DoWork(transaction); err != nil {
		if err := event.Nack(true); err != nil {
			fmt.Printf("ERROR: Nack() failed: %v\n", err)
		}
	} else {
		if err := event.Ack(); err != nil {
			fmt.Printf("ERROR: Ack() failed: %v\n", err)
		}
	}
}

func main() {
	eventbus, err := cores.NewAmqpClient("amqp://:5672", "example-consumer")
	if err != nil {
		panic(err)
	}

	eventbus.Subscribe("transaction-finished", false, new(Handler))

	// Block, since Subscibe() runs in its own goroutine
	select {}
}
