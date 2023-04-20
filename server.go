package bus

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	chanel *amqp.Channel
}

type replyMeta struct {
	replyTo       string
	correlationId string
}
type Msg struct {
	chanel    *amqp.Channel
	replyMeta replyMeta
	Delivery  *amqp.Delivery
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
				chanel: sr.chanel,
				replyMeta: replyMeta{
					replyTo:       m.ReplyTo,
					correlationId: m.CorrelationId,
				},
				Delivery: &m,
			}
		}
	}()

	return sr, nil
}

func (m *Msg) Reply(ctx context.Context, data []byte) error {
	err := m.chanel.PublishWithContext(ctx,
		"",                  // exchange
		m.replyMeta.replyTo, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			CorrelationId: m.replyMeta.correlationId,
			Body:          data,
		})

	return err
}
