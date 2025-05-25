package k9amqp

import (
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type amqpMetrics struct {
	PublishSent       *metrics.Metric
	PublishFailed     *metrics.Metric
	PublishLatency    *metrics.Metric
	ConsumeReceived   *metrics.Metric
	ConsumeNoDelivery *metrics.Metric
	ConsumeFailed     *metrics.Metric
	ConsumeLatency    *metrics.Metric
}

func registerMetrics(vu modules.VU) (amqpMetrics, error) {
	var err error
	registry := vu.InitEnv().Registry
	m := amqpMetrics{}
	m.PublishSent, err = registry.NewMetric("amqp_pub_sent", metrics.Counter)
	if err != nil {
		return m, err
	}
	m.PublishFailed, err = registry.NewMetric("amqp_pub_failed", metrics.Counter)
	if err != nil {
		return m, err
	}
	m.PublishFailed, err = registry.NewMetric("amqp_pub_latency", metrics.Trend)
	if err != nil {
		return m, err
	}
	m.ConsumeReceived, err = registry.NewMetric("amqp_sub_received", metrics.Counter)
	if err != nil {
		return m, err
	}
	m.ConsumeNoDelivery, err = registry.NewMetric("amqp_sub_no_delivery", metrics.Counter)
	if err != nil {
		return m, err
	}
	m.ConsumeFailed, err = registry.NewMetric("amqp_sub_failed", metrics.Counter)
	if err != nil {
		return m, err
	}
	m.PublishFailed, err = registry.NewMetric("amqp_sub_latency", metrics.Trend)
	if err != nil {
		return m, err
	}
	return m, nil

}
