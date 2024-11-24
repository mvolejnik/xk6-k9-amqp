package k9amqp

import (
	"errors"
	"log/slog"
)

func (queue *Exchange) Declare(client *Client, opts ExchangeDeclareOptions) error {
	var err error
	if client == nil {
		return errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
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

func (queue *Exchange) Delete(client *Client, opts ExchangeDeleteOptions) error {
	var err error
	if client == nil {
		return errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
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

func (exchange *Exchange) Bind(client *Client, opts ExchangeBindOptions) error {
	if client == nil {
		return errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
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

func (exchange *Exchange) Unind(client *Client, opts ExchangeUnbindOptions) error {
	if client == nil {
		return errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
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
