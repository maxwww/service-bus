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

func NewServer(conn *amqp.Connection, queueName string, router *Router) (*server, error) {
	chanel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = chanel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	replyToMsgs, err := chanel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	sr := &server{
		chanel: chanel,
	}

	go func() {
		for m := range replyToMsgs {
			go router.ServeRMQ(&Msg{
				chanel: sr.chanel,
				replyMeta: replyMeta{
					replyTo:       m.ReplyTo,
					correlationId: m.CorrelationId,
				},
				Delivery: &m,
			})
		}
	}()

	return sr, nil
}

func (m *Msg) Reply(ctx context.Context, data []byte) error {
	err := m.chanel.PublishWithContext(ctx,
		"",
		m.replyMeta.replyTo,
		false,
		false,
		amqp.Publishing{
			CorrelationId: m.replyMeta.correlationId,
			Body:          data,
		})

	return err
}
