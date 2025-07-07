package k9amqp

import amqp "github.com/rabbitmq/amqp091-go"

type (
	AmqpOptions struct {
		Host     string
		Port     int
		Vhost    string
		Username string
		Password string
	}

	PoolOptions struct {
		ChannelsPerConn   int
		ChannelsCacheSize int
	}

	Queue struct {
		amqpClient *AmqpClient
	}

	Exchange struct {
		amqpClient *AmqpClient
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

	QueueUnbindOptions struct {
		Name, Key, Exchange string
		Args                amqp.Table
	}

	QueuePurgeOptions struct {
		Name   string
		NoWait bool
	}

	PublishOptions struct {
		Exchange, Key        string
		Mandatory, Immediate bool
	}

	GetOptions struct {
		Queue   string
		AutoAck bool
	}

	ConsumeOptions struct {
		Queue     string
		AutoAck   bool
		Exclusive bool
		NoLocal   bool
		NoWait    bool
		Args      amqp.Table
		Size      int
	}

	AmqpProduceResponse struct {
		Error        bool
		ErrorMessage string
	}

	AmqpGetResponse struct {
		Delivery     amqp.Delivery
		Ok           bool
		Error        bool
		ErrorMessage string
	}

	AmqpConsumeResponse struct {
		Deliveries   []amqp.Delivery
		Ok           bool
		Error        bool
		ErrorMessage string
	}

	ListenerType func(amqp.Delivery) error

	ListenOptions struct {
		Queue     string
		AutoAck   bool
		Exclusive bool
		NoLocal   bool
		NoWait    bool
		Args      amqp.Table
	}
)
