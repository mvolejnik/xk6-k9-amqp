package k9amqp

import (
	"errors"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (queue *Queue) Declare(client *Client, opts QueueDeclareOptions) (*amqp.Queue, error) {
	var err error
	var amqpQueue amqp.Queue
	if client == nil {
		return nil, errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return nil, err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
		} else {
			slog.Info("blows channel after error")
		}
	}(err)
	if opts.Passive {
		amqpQueue, err = channel.QueueDeclarePassive(
			opts.Name,
			opts.Durable,
			opts.AutoDelete,
			opts.Exclusive,
			opts.NoWait,
			opts.Args,
		)
	} else {
		amqpQueue, err = channel.QueueDeclare(
			opts.Name,
			opts.Durable,
			opts.AutoDelete,
			opts.Exclusive,
			opts.NoWait,
			opts.Args,
		)
	}
	if err != nil {
		return nil, err
	}
	slog.Info("qeuue created", "name", amqpQueue.Name)
	return &amqpQueue, nil
}

func (queue *Queue) Delete(client *Client, opts QueueDeleteOptions) error {
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
	_, err = channel.QueueDelete(
		opts.Name,
		opts.IfUnused,
		opts.IfEmpty,
		opts.NoWait,
	)
	if err != nil {
		return err
	}
	slog.Info("qeuue deleted", "name", opts.Name)
	return nil
}

func (queue *Queue) Bind(client *Client, opts QueueBindOptions) error {
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
	err = channel.QueueBind(
		opts.Name,
		opts.Key,
		opts.Exchange,
		opts.NoWait,
		opts.Args,
	)
	slog.Info("qeuue binded", "name", opts.Name, "key", opts.Key)
	return err
}

func (queue *Queue) Unbind(client *Client, opts QueueUnbindOptions) error {
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
	err = channel.QueueUnbind(
		opts.Name,
		opts.Key,
		opts.Exchange,
		opts.Args,
	)
	slog.Info("qeuue unbinded", "name", opts.Name, "key", opts.Key)
	return err
}

func (queue *Queue) Purge(client *Client, opts QueuePurgeOptions) (int, error) {
	if client == nil {
		return 0, errors.New("required 'client' parameter missing")
	}
	channel, err := client.amqpClient.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return 0, err
	}
	defer func(err error) {
		if err == nil {
			client.amqpClient.channels.put(channel, err)
		} else {
			slog.Info("blows channel after error")
		}
	}(err)
	count, err := channel.QueuePurge(
		opts.Name,
		opts.NoWait,
	)
	slog.Info("qeuue purged", "name", opts.Name, "messages", count)
	return count, err
}
