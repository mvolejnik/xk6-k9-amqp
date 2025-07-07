package k9amqp

import (
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpClient struct {
	amqpOptions AmqpOptions
	poolOptions PoolOptions
	connections []*amqp.Connection
	channels    AmqpPool
}

type AmqpPool struct {
	connections []*amqp.Connection
	pool        chan *amqp.Channel
	mutex       sync.Mutex
	nextId      int
	connIdx     int
}

func (opt *AmqpOptions) init() {
	if opt.Host == "" {
		opt.Host = "localhost"
	}
	if opt.Port == 0 {
		opt.Port = 5672
	}
	if opt.Vhost == "" {
		opt.Vhost = "/"
	}
	if opt.Username == "" {
		opt.Username = "guest"
	}
	if opt.Password == "" {
		opt.Password = "guest"
	}
}

func (opt *PoolOptions) init() {
	if opt.ChannelsPerConn <= 0 {
		opt.ChannelsPerConn = 2
	}
	if opt.ChannelsCacheSize < 0 {
		opt.ChannelsCacheSize = 1
	}
}

func (p *AmqpPool) get() (*amqp.Channel, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	select {
	case channel := <-p.pool:
		if channel.IsClosed() {
			slog.Info("channel from pool is closed, creating new one")
			return p.channel()
		}
		return channel, nil
	default:
		slog.Info("no available channel in pool, creating new one")
		return p.channel()
	}
}

func (p *AmqpPool) channel() (*amqp.Channel, error) {
	p.nextId++
	p.connIdx = ((p.connIdx + 1) % len(p.connections))
	channel, err := p.connections[p.connIdx].Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (p *AmqpPool) put(channel *amqp.Channel, amqpError error) error {
	if channel.IsClosed() {
		slog.Warn("blow channel after error")
		return nil
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	select {
	case p.pool <- channel:
		return nil
	default:
		slog.Info("pool is full, closing channel")
		return channel.Close()
	}
}

func (amqpClient *AmqpClient) init() error {
	amqpClient.amqpOptions.init()
	amqpClient.poolOptions.init()
	if amqpClient.poolOptions.ChannelsPerConn == 0 {
		amqpClient.poolOptions.ChannelsPerConn = 2
	}
	var connSize = amqpClient.poolOptions.ChannelsCacheSize / amqpClient.poolOptions.ChannelsPerConn
	if connSize <= 0 {
		connSize = 1
	}
	var connections = make([]*amqp.Connection, connSize)
	for idx := range connSize {
		conn, err := amqpClient.Connect()
		if err != nil {
			return err
		}
		connections[idx] = conn
	}
	amqpClient.connections = connections
	amqpClient.channels = AmqpPool{
		connections: connections,
		pool:        make(chan *amqp.Channel, amqpClient.poolOptions.ChannelsCacheSize)}
	return nil
}

func (amqpClient *AmqpClient) close() error {
	for _, conn := range amqpClient.connections {
		conn.Close()
	}
	return nil
}

func (amqpClient *AmqpClient) Connect() (*amqp.Connection, error) {
	amqpConfig := amqp.Config{
		Vhost:      amqpClient.amqpOptions.Vhost,
		Properties: amqp.NewConnectionProperties(),
	}
	var amqpUrl = fmt.Sprintf("amqp://%s:%s@%s:%d", amqpClient.amqpOptions.Username, amqpClient.amqpOptions.Password, amqpClient.amqpOptions.Host, amqpClient.amqpOptions.Port)
	conn, err := amqp.DialConfig(amqpUrl, amqpConfig)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return conn, nil
}
