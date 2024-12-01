package k9amqp

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/grafana/sobek"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

var amqpClient *AmqpClient
var once sync.Once

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type K9amqp struct {
	vu      modules.VU
	metrics amqpMetrics
}

type Client struct {
	amqpClient AmqpClient
	k9amqp     K9amqp
}

func (k9amqp *K9amqp) XClient(call sobek.ConstructorCall, rt *sobek.Runtime) *sobek.Object {
	amqpOptions := new(AmqpOptions)
	poolOptions := new(PoolOptions)
	if len(call.Arguments) >= 1 {
		rt.ExportTo(call.Arguments[0], amqpOptions)
	}
	if len(call.Arguments) >= 2 {
		rt.ExportTo(call.Arguments[1], poolOptions)
	}
	amqpOptions.init()
	poolOptions.init()
	client := &Client{k9amqp: *k9amqp}
	client.init(*amqpOptions, *poolOptions)
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
	client.amqpClient = *amqpClient
	return err
}

func (client *Client) Teardown() {
	slog.Info("Teardown AMQP Client")
	client.amqpClient.close()
}

func (client *Client) Publish(opts PublishOptions, msg amqp.Publishing) (AmqpProduceResponse, error) {
	var err error
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpProduceResponse{Error: true, ErrorMessage: err.Error()}, err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
		} else {
			slog.Info("blows channel after error")
		}
	}(err)
	err = channel.Publish(
		opts.Exchange,
		opts.Key,
		opts.Mandatory,
		opts.Immediate,
		msg,
	)
	var errMessage string
	if err != nil {
		errMessage = err.Error()
	}
	response := AmqpProduceResponse{Error: err != nil, ErrorMessage: errMessage}
	client.k9amqp.reportPublishMetrics(*client, response)
	if err != nil {
		return response, err
	}
	return response, nil
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
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
		} else {
			slog.Info("blows channel after error")
		}
	}(err)
	delivery, ok, err = channel.Get(
		opts.Queue,
		opts.AutoAck,
	)
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	response := AmqpGetResponse{Delivery: delivery, Ok: ok, Error: err != nil, ErrorMessage: errorMessage}
	client.k9amqp.reportGetMetrics(*client, response)
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
	defer func(err error) {
		channel.Cancel(consumerTag, opts.NoWait)
		if err == nil {
			client.amqpClient.channels.put(channel, err)
		} else {
			slog.Info("blows channel after error")
		}
	}(err)
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
	client.k9amqp.reportConsumeMetrics(*client, response)
	if err != nil || len(deliveries) == 0 {
		return response, err
	}
	return response, nil
}

func (k9amqp *K9amqp) reportPublishMetrics(client Client, resp AmqpProduceResponse) error {
	now := time.Now()
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	tags := ctm.Tags.With("endpoint", fmt.Sprintf("amqp://%s@%s:%d/%s", client.amqpClient.amqpOptions.Username, client.amqpClient.amqpOptions.Host, client.amqpClient.amqpOptions.Port, client.amqpClient.amqpOptions.Vhost))
	ctx := k9amqp.vu.Context()
	var sent int
	var failed int
	if resp.Error {
		failed = 1
	} else {
		sent = 1
	}
	metrics.PushIfNotDone(ctx, k9amqp.vu.State().Samples, metrics.ConnectedSamples{
		Samples: []metrics.Sample{
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.PublishSent,
					Tags:   tags,
				},
				Value:    float64(sent),
				Metadata: ctm.Metadata,
			},
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.PublishFailed,
					Tags:   tags,
				},
				Value:    float64(failed),
				Metadata: ctm.Metadata,
			},
		},
	})
	return nil
}

func (k9amqp *K9amqp) reportGetMetrics(client Client, resp AmqpGetResponse) error {
	now := time.Now()
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	tags := ctm.Tags.With("endpoint", fmt.Sprintf("amqp://%s@%s:%d/%s", client.amqpClient.amqpOptions.Username, client.amqpClient.amqpOptions.Host, client.amqpClient.amqpOptions.Port, client.amqpClient.amqpOptions.Vhost))
	ctx := k9amqp.vu.Context()
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
	metrics.PushIfNotDone(ctx, k9amqp.vu.State().Samples, metrics.ConnectedSamples{
		Samples: []metrics.Sample{
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeReceived,
					Tags:   tags,
				},
				Value:    float64(received),
				Metadata: ctm.Metadata,
			},
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeNoDelivery,
					Tags:   tags,
				},
				Value:    float64(noDelivery),
				Metadata: ctm.Metadata,
			},
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeFailed,
					Tags:   tags,
				},
				Value:    float64(failed),
				Metadata: ctm.Metadata,
			},
		},
	})
	return nil
}

func (k9amqp *K9amqp) reportConsumeMetrics(client Client, resp AmqpConsumeResponse) error {
	now := time.Now()
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	tags := ctm.Tags.With("endpoint", fmt.Sprintf("amqp://%s@%s:%d/%s", client.amqpClient.amqpOptions.Username, client.amqpClient.amqpOptions.Host, client.amqpClient.amqpOptions.Port, client.amqpClient.amqpOptions.Vhost))
	ctx := k9amqp.vu.Context()
	var noDelivery int
	var failed int
	received := len(resp.Deliveries)
	if resp.Error {
		failed = 1
	} else if !resp.Ok {
		noDelivery = 1
	}
	metrics.PushIfNotDone(ctx, k9amqp.vu.State().Samples, metrics.ConnectedSamples{
		Samples: []metrics.Sample{
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeReceived,
					Tags:   tags,
				},
				Value:    float64(received),
				Metadata: ctm.Metadata,
			},
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeNoDelivery,
					Tags:   tags,
				},
				Value:    float64(noDelivery),
				Metadata: ctm.Metadata,
			},
			{
				Time: now,
				TimeSeries: metrics.TimeSeries{
					Metric: k9amqp.metrics.ConsumeFailed,
					Tags:   tags,
				},
				Value:    float64(failed),
				Metadata: ctm.Metadata,
			},
		},
	})
	return nil
}

func randString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
