package bus

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type client struct {
	chanel           *amqp.Channel
	serviceQueueName string
	replyQueueName   string
	senders          map[string]chan []byte
	mutex            *sync.Mutex
}

func NewClient(conn *amqp.Connection, queueName string) (*client, error) {
	chanel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	//serviceQueue, err := chanel.QueueDeclare(
	_, err = chanel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			//"x-message-ttl": 60000,
		}, // arguments
	)
	if err != nil {
		return nil, err
	}

	replyToQueue, err := chanel.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	replyToMsgs, err := chanel.Consume(
		replyToQueue.Name, // queue
		"",                // consumer
		true,              // auto-ack
		true,              // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)

	cl := &client{
		chanel:           chanel,
		serviceQueueName: queueName,
		replyQueueName:   replyToQueue.Name,
		senders:          make(map[string]chan []byte),
		mutex:            new(sync.Mutex),
	}

	go func() {
		for m := range replyToMsgs {
			cl.mutex.Lock()
			if cn, ok := cl.senders[m.CorrelationId]; ok {
				cn <- m.Body
				delete(cl.senders, m.CorrelationId)
			}
			cl.mutex.Unlock()
		}
	}()

	return cl, nil
}

func (b *client) Send(ctx context.Context, path string, msg []byte) ([]byte, error) {
	rCh := make(chan []byte)
	id := uuid.New().String()

	b.mutex.Lock()
	b.senders[id] = rCh
	b.mutex.Unlock()

	defer func() {
		b.mutex.Lock()
		delete(b.senders, id)
		close(rCh)
		b.mutex.Unlock()
	}()

	err := b.chanel.PublishWithContext(ctx,
		"",                 // exchange
		b.serviceQueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			Headers: amqp.Table{
				"path": path,
			},
			Expiration:    "60000",
			ReplyTo:       b.replyQueueName,
			CorrelationId: id,
			Body:          msg,
		})
	if err != nil {
		return nil, err
	}

	var response []byte
	select {
	case response = <-rCh:
	case <-ctx.Done():
		return nil, fmt.Errorf("declined by client")
	}

	return response, nil
}

func (b *client) Emit(ctx context.Context, msg []byte) error {

	return nil
}
