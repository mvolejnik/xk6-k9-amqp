package k9amqp

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/grafana/sobek"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

var amqpClient *AmqpClient
var once sync.Once

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

func (client *Client) Get(opts GetDeliveryOptions) (AmqpConsumeResponse, error) {
	var err error
	var delivery amqp.Delivery
	var ok bool
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpConsumeResponse{Error: true}, err
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
	response := AmqpConsumeResponse{Delivery: delivery, Ok: ok, Error: err != nil, ErrorMessage: errorMessage}
	client.k9amqp.reportConsumeMetrics(*client, response)
	if err != nil || !ok {
		return AmqpConsumeResponse{Error: true}, err
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

func (k9amqp *K9amqp) reportConsumeMetrics(client Client, resp AmqpConsumeResponse) error {
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
