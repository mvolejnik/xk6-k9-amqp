package k9amqp

import amqp "github.com/rabbitmq/amqp091-go"

type Queue struct {
	client *AmqpClient
}

type Exchange struct {
	client *AmqpClient
}

type ExchangeDeclareOptions struct {
	Name, Kind                            string
	Durable, AutoDelete, Internal, NoWait bool
	Args                                  amqp.Table
}

type ExchangeBindOptions struct {
	Destination, Source, Key string
	NoWait                   bool
	Args                     amqp.Table
}

type ExchangeUnbindOptions struct {
	Destination, Source, Key string
	NoWait                   bool
	Args                     amqp.Table
}

type ExchangeDeleteOptions struct {
	Name             string
	IfUnused, NoWait bool
}

type QueueDeclareOptions struct {
	Name                                            string
	Durable, AutoDelete, Exclusive, NoWait, Passive bool
	Args                                            amqp.Table
}

type QueueDeleteOptions struct {
	Name                      string
	IfUnused, IfEmpty, NoWait bool
}

type QueueBindOptions struct {
	Name, Key, Exchange string
	NoWait              bool
	Args                amqp.Table
}

type QueueUnindOptions struct {
	Name, Key, Exchange string
	Args                amqp.Table
}

type PublishOptions struct {
	Exchange, Key        string
	Mandatory, Immediate bool
}

type GetDeliveryOptions struct {
	Queue   string
	AutoAck bool
}
