package bus

import (
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
