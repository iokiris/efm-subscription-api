package service

import (
	"fmt"

	"github.com/iokiris/efm-subscription-api/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// Publisher интерфейс для сервисного слоя
type Publisher interface {
	Publish(exchange, routingKey string, body []byte) error
	Close()
}

type RabbitPublisher struct {
	ch    *amqp.Channel
	queue chan publishMsg
}

type publishMsg struct {
	exchange, routingKey string
	body                 []byte
}

// NewRabbitPublisher создает экземпляр Publisher с буфером и воркером
func NewRabbitPublisher(ch *amqp.Channel, buffer int) *RabbitPublisher {
	r := &RabbitPublisher{
		ch:    ch,
		queue: make(chan publishMsg, buffer),
	}
	go r.worker()
	return r
}

func (r *RabbitPublisher) worker() {
	for msg := range r.queue {
		err := r.ch.Publish(
			msg.exchange,
			msg.routingKey,
			false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        msg.body,
			},
		)
		if err != nil {
			logger.L.Error("RabbitPublisher publish error:", zap.Error(err))
		}
	}
}

// Publish помещает сообщение в очередь
func (r *RabbitPublisher) Publish(exchange, routingKey string, body []byte) error {
	select {
	case r.queue <- publishMsg{exchange, routingKey, body}:
		return nil
	default:
		return fmt.Errorf("publisher queue full")
	}
}

// Close закрывает очередь
func (r *RabbitPublisher) Close() {
	close(r.queue)
}
