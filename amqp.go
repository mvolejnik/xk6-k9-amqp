package k9amqp

import (
	"context"
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
	listenMutex   sync.Mutex
	listenConns   map[*amqp.Connection]struct{}
}

type AmqpPool struct {
	connections []*amqp.Connection
	pool        chan *amqp.Channel
	// slots limits the number of open channels (idle + in use) in bounded mode, nil otherwise
	slots   chan struct{}
	connIdx atomic.Uint64
	confirm bool
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

// get returns an idle channel from the pool or opens a new one. In bounded mode it waits
// for a free channel when the limit is reached (until ctx is done).
func (p *AmqpPool) get(ctx context.Context) (*amqp.Channel, error) {
	for {
		select {
		case channel := <-p.pool:
			if channel.IsClosed() {
				slog.Debug("channel from pool is closed, dropping it")
				continue
			}
			return channel, nil
		default:
		}
		if p.slots == nil {
			slog.Debug("no available channel in pool, creating new one")
			return p.channel()
		}
		if ctx == nil {
			ctx = context.Background()
		}
		select {
		case channel := <-p.pool:
			if channel.IsClosed() {
				slog.Debug("channel from pool is closed, dropping it")
				continue
			}
			return channel, nil
		case p.slots <- struct{}{}:
			slog.Debug("no available channel in pool, creating new one")
			channel, err := p.channel()
			if err != nil {
				<-p.slots
				return nil, err
			}
			// frees the slot whenever the channel gets closed (error, pool full, teardown)
			closed := channel.NotifyClose(make(chan *amqp.Error, 1))
			go func() {
				for range closed {
				}
				<-p.slots
			}()
			return channel, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (p *AmqpPool) channel() (*amqp.Channel, error) {
	connIdx := p.connIdx.Add(1) % uint64(len(p.connections))
	channel, err := p.connections[connIdx].Channel()
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
	select {
	case p.pool <- channel:
		return nil
	default:
		slog.Debug("pool is full, closing channel")
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
	if amqpClient.poolOptions.Bounded {
		if amqpClient.poolOptions.ChannelsCacheSize > 0 {
			amqpClient.channels.slots = make(chan struct{}, amqpClient.poolOptions.ChannelsCacheSize)
		} else {
			slog.Warn("bounded pool requires channels_cache_size > 0, pool isn't bounded")
		}
	}
	amqpClient.endpoint = fmt.Sprintf("amqp://%s@%s:%d/%s", amqpClient.amqpOptions.Username, amqpClient.amqpOptions.Host, amqpClient.amqpOptions.Port, amqpClient.amqpOptions.Vhost)
	return nil
}

func (amqpClient *AmqpClient) addListenConn(conn *amqp.Connection) {
	amqpClient.listenMutex.Lock()
	defer amqpClient.listenMutex.Unlock()
	if amqpClient.listenConns == nil {
		amqpClient.listenConns = make(map[*amqp.Connection]struct{})
	}
	amqpClient.listenConns[conn] = struct{}{}
}

func (amqpClient *AmqpClient) removeListenConn(conn *amqp.Connection) {
	amqpClient.listenMutex.Lock()
	defer amqpClient.listenMutex.Unlock()
	delete(amqpClient.listenConns, conn)
}

// close closes pooled connections and connections of still running listeners.
func (amqpClient *AmqpClient) close() error {
	var closeErr error
	amqpClient.listenMutex.Lock()
	connections := append([]*amqp.Connection{}, amqpClient.connections...)
	for conn := range amqpClient.listenConns {
		connections = append(connections, conn)
	}
	amqpClient.listenMutex.Unlock()
	for _, conn := range connections {
		if conn.IsClosed() {
			continue
		}
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

// deliveryAcker is the Acknowledger of deliveries received with auto_ack disabled.
// It keeps the channel out of the pool until all its deliveries are acked/nacked/rejected,
// then returns the channel to the pool. Unsettled deliveries are requeued by the broker
// when the connection is closed (Teardown).
type deliveryAcker struct {
	channel *amqp.Channel
	pool    *AmqpPool
	mutex   sync.Mutex
	pending map[uint64]struct{}
	done    bool
}

func newDeliveryAcker(channel *amqp.Channel, pool *AmqpPool, deliveries ...*amqp.Delivery) {
	acker := &deliveryAcker{channel: channel, pool: pool, pending: make(map[uint64]struct{}, len(deliveries))}
	for _, d := range deliveries {
		acker.pending[d.DeliveryTag] = struct{}{}
		d.Acknowledger = acker
	}
}

func (a *deliveryAcker) Ack(tag uint64, multiple bool) error {
	err := a.channel.Ack(tag, multiple)
	a.settle(tag, multiple, err)
	return err
}

func (a *deliveryAcker) Nack(tag uint64, multiple, requeue bool) error {
	err := a.channel.Nack(tag, multiple, requeue)
	a.settle(tag, multiple, err)
	return err
}

func (a *deliveryAcker) Reject(tag uint64, requeue bool) error {
	err := a.channel.Reject(tag, requeue)
	a.settle(tag, false, err)
	return err
}

func (a *deliveryAcker) settle(tag uint64, multiple bool, err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.done {
		return
	}
	if err != nil {
		slog.Info("blows channel after ack error", "error", err)
		a.done = true
		if closeErr := a.channel.Close(); closeErr != nil {
			slog.Error("failed to close channel after ack error", "error", closeErr)
		}
		return
	}
	if multiple {
		for pendingTag := range a.pending {
			if pendingTag <= tag {
				delete(a.pending, pendingTag)
			}
		}
	} else {
		delete(a.pending, tag)
	}
	if len(a.pending) == 0 {
		a.done = true
		if putErr := a.pool.put(a.channel, nil); putErr != nil {
			slog.Error("failed to return channel to pool", "error", putErr)
		}
	}
}
