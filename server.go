package bus

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	chanel *amqp.Channel
}

type ReplyMeta struct {
	ReplyTo       string
	CorrelationId string
}

type Msg struct {
	ReplyMeta
	Body []byte
}

func NewServer(conn *amqp.Connection, queueName string, handler chan Msg) (*server, error) {
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
			handler <- Msg{
				ReplyMeta: ReplyMeta{
					ReplyTo:       m.ReplyTo,
					CorrelationId: m.CorrelationId,
				},
				Body: m.Body,
			}
		}
	}()

	return sr, nil
}

func (s *server) Reply(ctx context.Context, replyTo ReplyMeta, data []byte) error {
	err := s.chanel.PublishWithContext(ctx,
		"",              // exchange
		replyTo.ReplyTo, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			CorrelationId: replyTo.CorrelationId,
			Body:          data,
		})

	return err
}
