package k9amqp

import (
	"errors"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (queue *Queue) Declare(k9amqp *K9amqp, opts QueueDeclareOptions) (*amqp.Queue, error) {
	var err error
	var amqpQueue amqp.Queue
	if k9amqp == nil || !k9amqp.inited {
		return nil, errors.New("required 'k9amqp' parameter missing or not initialized")
	}
	channel, err := k9amqp.client.channels.get()
	if err != nil {
		slog.Error("unable to get amqp channel")
		return nil, err
	}
	defer func(err error) {
		if err == nil {
			k9amqp.client.channels.put(channel, err)
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

func (queue *Queue) Delete(k9amqp *K9amqp, opts QueueDeleteOptions) error {
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

func (queue *Queue) Bind(k9amqp *K9amqp, opts QueueBindOptions) error {
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

func (queue *Queue) Unbind(k9amqp *K9amqp, opts QueueBindOptions) error {
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
	err = channel.QueueBind(
		opts.Name,
		opts.Key,
		opts.Exchange,
		opts.NoWait,
		opts.Args,
	)
	slog.Info("qeuue unbinded", "name", opts.Name, "key", opts.Key)
	return err
}
