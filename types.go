package k9amqp

import amqp "github.com/rabbitmq/amqp091-go"

type (
	Queue struct {
		client *AmqpClient
	}

	Exchange struct {
		client *AmqpClient
	}

	ExchangeDeclareOptions struct {
		Name, Kind                            string
		Durable, AutoDelete, Internal, NoWait bool
		Args                                  amqp.Table
	}

	ExchangeBindOptions struct {
		Destination, Source, Key string
		NoWait                   bool
		Args                     amqp.Table
	}

	ExchangeUnbindOptions struct {
		Destination, Source, Key string
		NoWait                   bool
		Args                     amqp.Table
	}

	ExchangeDeleteOptions struct {
		Name             string
		IfUnused, NoWait bool
	}

	QueueDeclareOptions struct {
		Name                                            string
		Durable, AutoDelete, Exclusive, NoWait, Passive bool
		Args                                            amqp.Table
	}

	QueueDeleteOptions struct {
		Name                      string
		IfUnused, IfEmpty, NoWait bool
	}

	QueueBindOptions struct {
		Name, Key, Exchange string
		NoWait              bool
		Args                amqp.Table
	}

	QueueUnindOptions struct {
		Name, Key, Exchange string
		Args                amqp.Table
	}

	PublishOptions struct {
		Exchange, Key        string
		Mandatory, Immediate bool
	}

	GetDeliveryOptions struct {
		Queue   string
		AutoAck bool
	}

	AmqpProduceResponse struct {
		Error        bool
		ErrorMessage string
	}

	AmqpConsumeResponse struct {
		Delivery     amqp.Delivery
		Ok           bool
		Error        bool
		ErrorMessage string
	}
)
