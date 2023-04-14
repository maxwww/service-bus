package bus

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	chanel *amqp.Channel
}

func newServer(conn *amqp.Connection, queueName string, handler chan []byte) (*server, error) {
	chanel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = chanel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	replyToMsgs, err := chanel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	sr := &server{
		chanel: chanel,
	}

	go func() {
		for m := range replyToMsgs {
			handler <- m.Body
		}
	}()

	return sr, nil
}
