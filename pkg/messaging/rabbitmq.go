package messaging

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchangeName = "saga.events"

type Bus struct {
	Conn *amqp.Connection
	Chan *amqp.Channel
}

func Connect(url string) (*Bus, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &Bus{Conn: conn, Chan: ch}, nil
}

func (b *Bus) Close() {
	if b == nil {
		return
	}
	if b.Chan != nil {
		_ = b.Chan.Close()
	}
	if b.Conn != nil {
		_ = b.Conn.Close()
	}
}

func (b *Bus) Publish(routingKey string, event Event) error {
	body, err := event.Encode()
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}

	return b.Chan.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Timestamp:   time.Now().UTC(),
		Body:        body,
	})
}

func (b *Bus) Consume(queue, routingKey string) (<-chan amqp.Delivery, error) {
	if _, err := b.Chan.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := b.Chan.QueueBind(queue, routingKey, exchangeName, false, nil); err != nil {
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	msgs, err := b.Chan.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	return msgs, nil
}

func MustConnect(url string) *Bus {
	for {
		bus, err := Connect(url)
		if err == nil {
			return bus
		}
		log.Printf("waiting for rabbitmq: %v", err)
		time.Sleep(3 * time.Second)
	}
}
