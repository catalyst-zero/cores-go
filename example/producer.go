package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	coreos "github.com/catalyst-zero/cores-go"
	"time"
)

func newHash(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

type TransactionFinishedEvent struct {
	Id string `json:"id"`
}

func NewTransactionFinishedEvent() TransactionFinishedEvent {
	return TransactionFinishedEvent{
		Id: newHash(fmt.Sprintf("%d", time.Now())),
	}
}

func main() {
	eventbus, err := coreos.NewAmqpClient(":5672")
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(10 * time.Second)

		payload := NewTransactionFinishedEvent()
		eventbus.Broadcast("transaction-finished", payload)
	}
}
