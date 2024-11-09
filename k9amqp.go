package k9amqp

import (
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type K9amqp struct {
	vu      modules.VU
	metrics amqpMetrics
	client  AmqpClient
	inited  bool
}

func (k9amqp *K9amqp) Init(amqpOptions AmqpOptions, poolOptions PoolOptions) error {
	slog.Info(fmt.Sprintf("init amqp client with pool %+v\n", poolOptions))
	k9amqp.client = AmqpClient{amqpOptions: amqpOptions, poolOptions: poolOptions}
	err := k9amqp.client.init()
	if err == nil {
		k9amqp.inited = true
	}
	return err
}

func (k9amqp *K9amqp) Teardown() {
	slog.Info("Teardown AMQP Client")
	k9amqp.client.close()
}

func (k9amqp *K9amqp) Publish(opts PublishOptions, msg amqp.Publishing) (AmqpProduceResponse, error) {
	var err error
	channel, err := k9amqp.client.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpProduceResponse{Error: true, ErrorMessage: err.Error()}, err
	}
	defer func(err error) {
		if err == nil {
			k9amqp.client.channels.put(channel, err)
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
	k9amqp.reportPublishMetrics(response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (k9amqp *K9amqp) Get(opts GetDeliveryOptions) (AmqpConsumeResponse, error) {
	var err error
	var delivery amqp.Delivery
	var ok bool
	channel, err := k9amqp.client.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return AmqpConsumeResponse{Error: true}, err
	}
	defer func(err error) {
		if err == nil {
			k9amqp.client.channels.put(channel, err)
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
	k9amqp.reportConsumeMetrics(response)
	if err != nil || !ok {
		return AmqpConsumeResponse{Error: true}, err
	}
	return response, nil
}

func (k9amqp *K9amqp) reportPublishMetrics(resp AmqpProduceResponse) error {
	now := time.Now()
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	tags := ctm.Tags.With("endpoint", fmt.Sprintf("amqp://%s@%s:%d/%s", k9amqp.client.amqpOptions.Username, k9amqp.client.amqpOptions.Host, k9amqp.client.amqpOptions.Port, k9amqp.client.amqpOptions.Vhost))
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

func (k9amqp *K9amqp) reportConsumeMetrics(resp AmqpConsumeResponse) error {
	now := time.Now()
	ctm := k9amqp.vu.State().Tags.GetCurrentValues()
	tags := ctm.Tags.With("endpoint", fmt.Sprintf("amqp://%s@%s:%d/%s", k9amqp.client.amqpOptions.Username, k9amqp.client.amqpOptions.Host, k9amqp.client.amqpOptions.Port, k9amqp.client.amqpOptions.Vhost))
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
