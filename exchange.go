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

func (exchange *Exchange) Unbind(client *Client, opts ExchangeUnbindOptions) error {
	if client == nil {
		return errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return err
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
