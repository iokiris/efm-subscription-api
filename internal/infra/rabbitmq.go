package infra

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitConfig struct {
	User     string
	Password string
	Host     string
	Port     int
}

func NewRabbitMQ(cfg *RabbitConfig) (*amqp.Connection, *amqp.Channel, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port,
	)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		"subscriptions", // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // noWait
		nil,             // args
	); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("rabbitmq exchange declare: %w", err)
	}

	return conn, ch, nil
}
