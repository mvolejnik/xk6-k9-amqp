package k9amqp

import (
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpOptions struct {
	Host     string
	Port     int
	Vhost    string
	Username string
	Password string
}

type PoolOptions struct {
	ChannelsPerConn  int
	ChannelCacheSize int
}

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
	if opt.ChannelCacheSize < 0 {
		opt.ChannelCacheSize = 1
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

func (client *AmqpClient) init() error {
	client.amqpOptions.init()
	client.poolOptions.init()
	if client.poolOptions.ChannelsPerConn == 0 {
		client.poolOptions.ChannelsPerConn = 2
	}
	var connSize = client.poolOptions.ChannelCacheSize / client.poolOptions.ChannelsPerConn
	if connSize <= 0 {
		connSize = 1
	}
	var connections = make([]*amqp.Connection, connSize)
	for idx := range connSize {
		conn, err := client.connect()
		if err != nil {
			return err
		}
		connections[idx] = conn
	}
	client.connections = connections
	client.channels = AmqpPool{
		connections: connections,
		pool:        make(chan *amqp.Channel, client.poolOptions.ChannelCacheSize)}
	return nil
}

func (client *AmqpClient) close() error {
	for _, conn := range client.connections {
		conn.Close()
	}
	return nil
}

func (client *AmqpClient) connect() (*amqp.Connection, error) {
	amqpConfig := amqp.Config{
		Vhost:      client.amqpOptions.Vhost,
		Properties: amqp.NewConnectionProperties(),
	}
	var amqpUrl = fmt.Sprintf("amqp://%s:%s@%s:%d", client.amqpOptions.Username, client.amqpOptions.Password, client.amqpOptions.Host, client.amqpOptions.Port)
	conn, err := amqp.DialConfig(amqpUrl, amqpConfig)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return conn, nil
}
