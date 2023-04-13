package bus

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type bus struct {
	conn           *amqp.Connection
	chanel         *amqp.Channel
	qName          string
	replyQueueName string
	senders        map[string]chan string
	mutex          *sync.Mutex
}

func NewBus(conn *amqp.Connection) *bus {
	return &bus{
		conn:  conn,
		mutex: new(sync.Mutex),
	}
}

func (b *bus) Send(ctx context.Context, msg []byte) ([]byte, error) {

	return nil, nil
}

func (b *bus) Emit(ctx context.Context, msg []byte) error {

	return nil
}
