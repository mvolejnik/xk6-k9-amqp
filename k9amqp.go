package k9amqp

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/grafana/sobek"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.k6.io/k6/v2/js/modules"
	"go.k6.io/k6/v2/metrics"
)

var amqpClient *AmqpClient

// ErrPublishNack is returned when broker negatively acknowledges published message (publisher confirms).
var ErrPublishNack = errors.New("message nacked by broker")
var once sync.Once

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type K9amqp struct {
	vu      modules.VU
	metrics amqpMetrics
}

type Client struct {
	amqpClient *AmqpClient
	k9amqp     K9amqp
}

func (k9amqp *K9amqp) XClient(call sobek.ConstructorCall, rt *sobek.Runtime) *sobek.Object {
	amqpOptions := new(AmqpOptions)
	poolOptions := new(PoolOptions)
	if len(call.Arguments) >= 1 {
		if err := rt.ExportTo(call.Arguments[0], amqpOptions); err != nil {
			panic(rt.NewTypeError("failed to export amqpOptions: %v", err))
		}
	}
	if len(call.Arguments) >= 2 {
		if err := rt.ExportTo(call.Arguments[1], poolOptions); err != nil {
			panic(rt.NewTypeError("failed to export poolOptions: %v", err))
		}
	}
	amqpOptions.init()
	poolOptions.init()
	client := &Client{k9amqp: *k9amqp}
	if err := client.init(*amqpOptions, *poolOptions); err != nil {
		panic(rt.NewGoError(err))
	}
	return rt.ToValue(client).ToObject(rt)
}

func (client *Client) getAmqpClient(amqpOptions AmqpOptions, poolOptions PoolOptions) (*AmqpClient, error) {
	var err error
	once.Do(func() {
		slog.Info(fmt.Sprintf("init amqp client with pool %+v\n", poolOptions))
		amqpClient = &AmqpClient{amqpOptions: amqpOptions, poolOptions: poolOptions}
		err = amqpClient.init()
	})
	return amqpClient, err
}

func (client *Client) init(amqpOptions AmqpOptions, poolOptions PoolOptions) error {
	var err error
	amqpClient, err := client.getAmqpClient(amqpOptions, poolOptions)
	client.amqpClient = amqpClient
	return err
}

func (client *Client) Teardown() {
	slog.Info("Teardown AMQP Client")
	client.k9amqp.reportDataMetrics(client)
	if err := client.amqpClient.close(); err != nil {
		slog.Error("failed to close amqp client(s)")
	}
}

func (client *Client) Publish(opts PublishOptions, msg amqp.Publishing) (AmqpProduceResponse, error) {
	var err error
	var duration time.Duration
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpProduceResponse{Error: true, ErrorMessage: err.Error()}, err
	}
	defer func() {
		// nack doesn't close the channel, it can be reused
		if err == nil || errors.Is(err, ErrPublishNack) {
			if putErr := client.amqpClient.channels.put(channel, err); putErr != nil {
				slog.Error("failed to return channel to pool", "error", putErr)
			}
		} else {
			slog.Info("blows channel after error")
			if closeErr := channel.Close(); closeErr != nil {
				slog.Error("failed to close channel after error", "error", closeErr)
			}
		}
	}()
	duration, err = client.publish(channel, opts, msg)
	var errMessage string
	if err != nil {
		errMessage = err.Error()
	}
	response := AmqpProduceResponse{Error: err != nil, ErrorMessage: errMessage}
	if metricsErr := client.k9amqp.reportPublishMetrics(client, opts, response, duration); metricsErr != nil {
		slog.Error("failed to report publish metrics", "error", metricsErr)
	}
	if err != nil {
		return response, err
	}
	return response, nil
}

// publish waits for broker confirmation when channel is in confirm mode,
// otherwise duration covers just writing the message to the socket.
func (client *Client) publish(channel *amqp.Channel, opts PublishOptions, msg amqp.Publishing) (time.Duration, error) {
	ctx := client.k9amqp.vu.Context()
	startTime := time.Now()
	confirmation, err := channel.PublishWithDeferredConfirmWithContext(
		ctx,
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		msg,
	)
	if err == nil && confirmation != nil {
		var acked bool
		acked, err = confirmation.WaitContext(ctx)
		if err == nil && !acked {
			err = ErrPublishNack
		}
	}
	duration := time.Since(startTime)
	return duration, err
}

func (client *Client) Get(opts GetOptions) (AmqpGetResponse, error) {
	var err error
	var delivery amqp.Delivery
	var ok bool
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpGetResponse{Error: true}, err
	}
	defer func() {
		if err == nil {
			if putErr := client.amqpClient.channels.put(channel, err); putErr != nil {
				slog.Error("failed to return channel to pool", "error", putErr)
			}
		} else {
			slog.Info("blows channel after error")
			if closeErr := channel.Close(); closeErr != nil {
				slog.Error("failed to close channel after error", "error", closeErr)
			}
		}
	}()
	delivery, ok, err = channel.Get(
		opts.Queue,
		opts.AutoAck,
	)
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	response := AmqpGetResponse{Delivery: delivery, Ok: ok, Error: err != nil, ErrorMessage: errorMessage}
	if metricsErr := client.k9amqp.reportGetMetrics(client, response); metricsErr != nil {
		slog.Error("failed to report get metrics", "error", metricsErr)
	}
	if err != nil || !ok {
		return response, err
	}
	return response, nil
}

func (client *Client) Consume(opts ConsumeOptions) (AmqpConsumeResponse, error) {
	var err error
	var consumerTag = randString(10)
	deliveries := []amqp.Delivery{}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpConsumeResponse{Error: true}, err
	}
	defer func() {
		if cancelErr := channel.Cancel(consumerTag, opts.NoWait); cancelErr != nil {
			slog.Error("failed to cancel consumer", "error", cancelErr)
		}
		if err == nil {
			if putErr := client.amqpClient.channels.put(channel, err); putErr != nil {
				slog.Error("failed to return channel to pool", "error", putErr)
			}
		} else {
			slog.Info("blows channel after error")
			if closeErr := channel.Close(); closeErr != nil {
				slog.Error("failed to close channel after error", "error", closeErr)
			}
		}
	}()
	amqpChannel, err := channel.Consume(
		opts.Queue,
		consumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
	var loop = true
	for range opts.Size {
		select {
		case d, ok := <-amqpChannel:
			if ok {
				deliveries = append(deliveries, d)
			}
		default:
			loop = false
		}
		if !loop {
			break
		}
	}
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	response := AmqpConsumeResponse{Deliveries: deliveries, Ok: len(deliveries) > 0, Error: err != nil, ErrorMessage: errorMessage}
	if metricsErr := client.k9amqp.reportConsumeMetrics(client, response); metricsErr != nil {
		slog.Error("failed to report consume metrics", "error", metricsErr)
	}
	if err != nil || len(deliveries) == 0 {
		return response, err
	}
	return response, nil
}

func (client *Client) Listen(opts ListenOptions, listener ListenerType) error {
	var err error
	var consumerTag = randString(10)
	conn, err := client.amqpClient.Connect()
	if err != nil {
		slog.Error("unable to get amqp connection")
		return err
	}
	channel, err := conn.Channel()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
	}
	amqpChannel, err := channel.Consume(
		opts.Queue,
		consumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		slog.Error("unable to consume from queue", "error", err)
		if closeErr := channel.Close(); closeErr != nil {
			slog.Error("failed to close channel", "error", closeErr)
		}
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("failed to close connection", "error", closeErr)
		}
		return err
	}

	go func() {
		defer func() {
			if cancelErr := channel.Cancel(consumerTag, opts.NoWait); cancelErr != nil {
				slog.Error("failed to cancel consumer", "error", cancelErr)
			}
			if closeErr := channel.Close(); closeErr != nil {
				slog.Error("failed to close channel", "error", closeErr)
			}
			if closeErr := conn.Close(); closeErr != nil {
				slog.Error("failed to close connection", "error", closeErr)
			}
		}()
		for d := range amqpChannel {
			if err := listener(d); err != nil {
				slog.Error(fmt.Sprintf("%s (Listen)", err.Error()))
				return
			}
		}
	}()
	return nil
}

func (k9amqp *K9amqp) reportPublishMetrics(client *Client, opts PublishOptions, resp AmqpProduceResponse, duration time.Duration) error {
	tags := k9amqp.endpointTags(client)
	if tags == nil {
		return nil
	}
	tags = tags.With("exchange", opts.Exchange)
	tags = tags.With("routing_key", opts.Key)
	var sent int
	var failed int
	if resp.Error {
		failed = 1
	} else {
		sent = 1
	}
	k9amqp.push(
		k9amqp.sample(k9amqp.metrics.PublishSent, tags, float64(sent)),
		k9amqp.sample(k9amqp.metrics.PublishFailed, tags, float64(failed)),
		k9amqp.sample(k9amqp.metrics.PublishDuration, tags, metrics.D(duration)),
	)
	k9amqp.reportDataMetrics(client)
	return nil
}

func (k9amqp *K9amqp) reportGetMetrics(client *Client, resp AmqpGetResponse) error {
	tags := k9amqp.endpointTags(client)
	if tags == nil {
		return nil
	}
	tags = k9amqp.metricsTags(tags, &resp.Delivery)
	var received int
	var noDelivery int
	var failed int
	if resp.Ok {
		received = 1
	} else if resp.Error {
		failed = 1
	} else if !resp.Ok {
		noDelivery = 1
	}
	k9amqp.push(
		k9amqp.sample(k9amqp.metrics.ConsumeReceived, tags, float64(received)),
		k9amqp.sample(k9amqp.metrics.ConsumeNoDelivery, tags, float64(noDelivery)),
		k9amqp.sample(k9amqp.metrics.ConsumeFailed, tags, float64(failed)),
	)
	k9amqp.reportDataMetrics(client)
	return nil
}

func (k9amqp *K9amqp) reportConsumeMetrics(client *Client, resp AmqpConsumeResponse) error {
	tags := k9amqp.endpointTags(client)
	if tags == nil {
		return nil
	}
	var delivery *amqp.Delivery
	if len(resp.Deliveries) > 0 {
		delivery = &resp.Deliveries[0]
	}
	tags = k9amqp.metricsTags(tags, delivery)
	var noDelivery int
	var failed int
	received := len(resp.Deliveries)
	if resp.Error {
		failed = 1
	} else if !resp.Ok {
		noDelivery = 1
	}
	k9amqp.push(
		k9amqp.sample(k9amqp.metrics.ConsumeReceived, tags, float64(received)),
		k9amqp.sample(k9amqp.metrics.ConsumeNoDelivery, tags, float64(noDelivery)),
		k9amqp.sample(k9amqp.metrics.ConsumeFailed, tags, float64(failed)),
	)
	k9amqp.reportDataMetrics(client)
	return nil
}

// reportDataMetrics pushes bytes sent/received on the wire by all AMQP connections
// since the last report into k6 builtin data_sent/data_received metrics.
// Counters are global (shared by all VUs), whichever VU reports first takes the delta.
func (k9amqp *K9amqp) reportDataMetrics(client *Client) {
	state := k9amqp.vu.State()
	if state == nil || client.amqpClient == nil {
		return
	}
	sent := client.amqpClient.bytesSent.Swap(0)
	received := client.amqpClient.bytesReceived.Swap(0)
	if sent == 0 && received == 0 {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags
	k9amqp.push(
		k9amqp.sample(state.BuiltinMetrics.DataSent, tags, float64(sent)),
		k9amqp.sample(state.BuiltinMetrics.DataReceived, tags, float64(received)),
	)
}

// endpointTags returns current VU tags with 'endpoint' tag, nil outside of VU context.
func (k9amqp *K9amqp) endpointTags(client *Client) *metrics.TagSet {
	state := k9amqp.vu.State()
	if state == nil {
		return nil
	}
	return state.Tags.GetCurrentValues().Tags.With("endpoint", client.amqpClient.endpoint)
}

func (k9amqp *K9amqp) sample(metric *metrics.Metric, tags *metrics.TagSet, value float64) metrics.Sample {
	return metrics.Sample{
		Time: time.Now(),
		TimeSeries: metrics.TimeSeries{
			Metric: metric,
			Tags:   tags,
		},
		Value:    value,
		Metadata: k9amqp.vu.State().Tags.GetCurrentValues().Metadata,
	}
}

func (k9amqp *K9amqp) push(samples ...metrics.Sample) {
	metrics.PushIfNotDone(k9amqp.vu.Context(), k9amqp.vu.State().Samples, metrics.ConnectedSamples{
		Samples: samples,
	})
}

func randString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (k9amqp *K9amqp) scenario() (*string, bool) {
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	if scenario, ok := ctm.Tags.Get("scenario"); ok {
		return &scenario, true
	}
	return nil, false
}

func (k9amqp *K9amqp) metricsTags(tags *metrics.TagSet, delivery *amqp.Delivery) *metrics.TagSet {
	scenario, scenaried := k9amqp.scenario()
	if scenaried {
		tags = tags.With("scenario", *scenario)
	}
	if delivery != nil {
		tags = tags.With("exchange", delivery.Exchange)
		tags = tags.With("routing_key", delivery.RoutingKey)
	}
	return tags
}
