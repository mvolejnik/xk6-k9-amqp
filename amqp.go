package k9amqp

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionTimeout = 30 * time.Second

type AmqpClient struct {
	amqpOptions   AmqpOptions
	poolOptions   PoolOptions
	connections   []*amqp.Connection
	channels      AmqpPool
	endpoint      string
	bytesSent     atomic.Int64
	bytesReceived atomic.Int64
}

type AmqpPool struct {
	connections []*amqp.Connection
	pool        chan *amqp.Channel
	mutex       sync.Mutex
	nextId      int
	connIdx     int
	confirm     bool
}

// countingConn counts bytes on the wire of all AMQP connections (global, not per VU).
type countingConn struct {
	net.Conn
	sent, received *atomic.Int64
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.received.Add(int64(n))
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.sent.Add(int64(n))
	return n, err
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
	if p.confirm {
		if err := channel.Confirm(false); err != nil {
			if closeErr := channel.Close(); closeErr != nil {
				slog.Error("failed to close channel", "error", closeErr)
			}
			return nil, err
		}
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
		pool:        make(chan *amqp.Channel, amqpClient.poolOptions.ChannelsCacheSize),
		confirm:     !amqpClient.poolOptions.NoConfirm}
	amqpClient.endpoint = fmt.Sprintf("amqp://%s@%s:%d/%s", amqpClient.amqpOptions.Username, amqpClient.amqpOptions.Host, amqpClient.amqpOptions.Port, amqpClient.amqpOptions.Vhost)
	return nil
}

func (amqpClient *AmqpClient) close() error {
	var closeErr error
	for _, conn := range amqpClient.connections {
		err := conn.Close()
		if closeErr == nil && err != nil {
			closeErr = err
		}
		if err != nil {
			slog.Error(err.Error())
		}
	}
	return closeErr
}

func (amqpClient *AmqpClient) Connect() (*amqp.Connection, error) {
	amqpConfig := amqp.Config{
		Vhost:      amqpClient.amqpOptions.Vhost,
		Properties: amqp.NewConnectionProperties(),
		Dial: func(network, addr string) (net.Conn, error) {
			conn, err := amqp.DefaultDial(connectionTimeout)(network, addr)
			if err != nil {
				return nil, err
			}
			return &countingConn{Conn: conn, sent: &amqpClient.bytesSent, received: &amqpClient.bytesReceived}, nil
		},
	}
	var amqpUrl = fmt.Sprintf("amqp://%s:%s@%s:%d", amqpClient.amqpOptions.Username, amqpClient.amqpOptions.Password, amqpClient.amqpOptions.Host, amqpClient.amqpOptions.Port)
	conn, err := amqp.DialConfig(amqpUrl, amqpConfig)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return conn, nil
}
