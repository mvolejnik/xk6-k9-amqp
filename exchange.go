package k9amqp

import (
	"errors"
	"log/slog"
)

func (queue *Exchange) Declare(k9amqp *K9amqp, opts ExchangeDeclareOptions) error {
	var err error
	if k9amqp == nil || !k9amqp.inited {
		return errors.New("required 'k9amqp' parameter missing or not initialized")
	}
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
	err = channel.ExchangeDeclare(
		opts.Name,
		opts.Kind,
		opts.Durable,
		opts.AutoDelete,
		opts.Internal,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return err
	}
	slog.Info("exchange created", "name", opts.Name)
	return nil
}

func (queue *Exchange) Delete(k9amqp *K9amqp, opts ExchangeDeleteOptions) error {
	var err error
	if k9amqp == nil || !k9amqp.inited {
		return errors.New("required 'k9amqp' parameter missing or not initialized")
	}
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
	err = channel.ExchangeDelete(
		opts.Name,
		opts.IfUnused,
		opts.NoWait,
	)
	if err != nil {
		return err
	}
	slog.Info("exchange deleted", "name", opts.Name)
	return nil
}

func (exchange *Exchange) Bind(k9amqp *K9amqp, opts ExchangeBindOptions) error {
	if k9amqp == nil || !k9amqp.inited {
		return errors.New("required 'k9amqp' parameter missing or not initialized")
	}
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
	err = channel.ExchangeBind(
		opts.Destination,
		opts.Key,
		opts.Source,
		opts.NoWait,
		opts.Args,
	)
	slog.Info("exchange binded", "destination", opts.Destination, "key", opts.Key, "source", opts.Source)
	return err
}

func (exchange *Exchange) Unind(k9amqp *K9amqp, opts ExchangeUnbindOptions) error {
	if k9amqp == nil || !k9amqp.inited {
		return errors.New("required 'k9amqp' parameter missing or not initialized")
	}
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
	err = channel.ExchangeUnbind(
		opts.Destination,
		opts.Key,
		opts.Source,
		opts.NoWait,
		opts.Args,
	)
	slog.Info("exchange binded", "destination", opts.Destination, "key", opts.Key, "source", opts.Source)
	return err
}
