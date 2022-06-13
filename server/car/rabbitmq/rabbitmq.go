package rabbitmq

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Publisher struct {
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(conn *amqp.Connection, exchange string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("cannot allocate channel: %v", err)
	}

	if err = declareExchange(ch, exchange); err != nil {
		return nil, fmt.Errorf("cannot declare exchange: %v", err)
	}
	return &Publisher{
		ch:       ch,
		exchange: exchange,
	}, nil
}

func (s Publisher) Publish(c context.Context, carEntity *carpb.CarEntity) error {
	bytes, err := json.Marshal(carEntity)
	if err != nil {
		return fmt.Errorf("cannot marshal: %v", err)
	}

	return s.ch.Publish(
		s.exchange,
		"",    // key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body: bytes,
		},
	)
}

// ==================================================================================================================

type Subscriber struct {
	conn     *amqp.Connection
	exchange string
	logger   *zap.Logger
}

func NewScriber(conn *amqp.Connection, exchange string, logger *zap.Logger) (*Subscriber, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("cannot allocate channel: %v", err)
	}
	defer func(ch *amqp.Channel) {
		_ = ch.Close()
	}(ch)

	if err = declareExchange(ch, exchange); err != nil {
		return nil, fmt.Errorf("cannot declare exchange: %v", err)
	}

	return &Subscriber{
		conn:     conn,
		exchange: exchange,
		logger:   logger,
	}, nil
}

func (s *Subscriber) SubscribeRaw(ctx context.Context) (<-chan amqp.Delivery, func(), error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot allocate channel: %v", err)
	}
	closeCh := func() {
		if err := ch.Close(); err != nil {
			s.logger.Error("cannot close cahnnel", zap.Error(err))
		}
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // autoDelete
		false, // exlusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, closeCh, fmt.Errorf("cannot declare queue: %v", err)
	}
	closeChQueue := func() {
		if _, err := ch.QueueDelete(q.Name, false, false, false); err != nil {
			s.logger.Error("cannot close queue:", zap.String("name", q.Name), zap.Error(err))
		}
		closeCh()
	}

	err = ch.QueueBind(
		q.Name,
		"", // key
		s.exchange,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, closeChQueue, fmt.Errorf("cannot bind: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, closeChQueue, fmt.Errorf("cannot consume queue: %v", err)
	}
	return msgs, closeChQueue, nil
}

func (s *Subscriber) Subscribe(ctx context.Context) (chan *carpb.CarEntity, func(), error) {
	rawCh, closeFunc, err := s.SubscribeRaw(ctx)
	if err != nil {
		return nil, closeFunc, err
	}

	carCh := make(chan *carpb.CarEntity)
	go func() {
		for rawMsg := range rawCh {
			var carEntity carpb.CarEntity
			err = json.Unmarshal(rawMsg.Body, &carEntity)
			if err != nil {
				s.logger.Error("cannot unmmarshal", zap.Error(err))
			}
			carCh <- &carEntity
		}
		close(carCh)
	}()
	return carCh, closeFunc, nil
}

func declareExchange(ch *amqp.Channel, exchange string) error {
	return ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)
}
