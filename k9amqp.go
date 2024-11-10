package k9amqp

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.k6.io/k6/js/modules"
)

type (
	K9amqp struct {
		vu     modules.VU
		client AmqpClient
		inited bool
	}

	RootModule     struct{}
	ModuleInstance struct {
		vu     modules.VU
		k9amqp *K9amqp
	}
)

var (
	_ modules.Instance = &ModuleInstance{}
	_ modules.Module   = &RootModule{}
)

func New() *RootModule {
	return &RootModule{}
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu:     vu,
		k9amqp: &K9amqp{vu: vu},
	}
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: mi.k9amqp,
	}
}

func init() {
	modules.Register("k6/x/k9amqp", New())
	modules.Register("k6/x/k9amqp/queue", new(Queue))
	modules.Register("k6/x/k9amqp/exchange", new(Exchange))
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

func (k9amqp *K9amqp) Publish(opts PublishOptions, msg amqp.Publishing) error {
	var err error
	channel, err := k9amqp.client.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
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
	if err != nil {
		return err
	}
	return nil
}

func (k9amqp *K9amqp) Get(opts GetDeliveryOptions) (amqp.Delivery, error) {
	var err error
	var delivery amqp.Delivery
	var ok bool
	channel, err := k9amqp.client.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return amqp.Delivery{}, err
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
	if err != nil || !ok {
		return amqp.Delivery{}, err
	}
	return delivery, nil
}
