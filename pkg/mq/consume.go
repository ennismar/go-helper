package mq

import (
	"context"
	"github.com/ennismar/go-helper/pkg/log"
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/houseofcat/turbocookedrabbit/v2/pkg/tcr"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"sync/atomic"
)

type Consume struct {
	ops      ConsumeOptions
	q        string
	consumer *tcr.Consumer
	Error    error
}

func (qu *Queue) Consume(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) (err error) {
	if handler == nil {
		err = errors.Errorf("handler is nil")
		return
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		err = errors.WithStack(co.Error)
		return
	}
	ctx := co.newContext(nil)
	co.consumer.StartConsumingWithAction(func(msg *tcr.ReceivedMessage) {
		d := msg.Delivery
		a := d.Acknowledger
		tag := d.DeliveryTag
		ok := handler(ctx, co.q, d)
		if co.ops.autoAck {
			return
		}
		if ok {
			e := a.Ack(tag, false)
			if e != nil {
				log.WithContext(ctx).WithError(e).Error("consume ack failed")
			}
			return
		}
		e := a.Nack(tag, false, co.ops.nackRequeue)
		if e != nil {
			log.WithContext(ctx).WithError(e).Error("consume nack failed")
		}
	})
	return
}

func (qu *Queue) ConsumeOne(size int, handler func(c context.Context, q string, d amqp.Delivery) bool, options ...func(*ConsumeOptions)) (err error) {
	if size < 1 {
		err = errors.Errorf("minimum size is 1")
		return
	}
	if handler == nil {
		err = errors.Errorf("handler is nil")
		return
	}

	if atomic.LoadInt32(&qu.ex.rb.lost) == 1 {
		err = errors.Errorf("connection maybe lost")
		return
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		err = errors.WithStack(co.Error)
		return
	}
	ch := qu.ex.rb.pool.GetChannelFromPool()
	defer func() {
		qu.ex.rb.pool.ReturnChannel(ch, true)
	}()
	var ds []amqp.Delivery
	ds, err = qu.getBatch(ch, qu.ops.name, size, co.ops.autoAck)

	if err != nil {
		err = errors.WithStack(err)
		return
	}
	ctx := co.newContext(co.ops.oneCtx)
	for i, d := range ds {
		a := d.Acknowledger
		tag := d.DeliveryTag
		ctx = context.WithValue(ctx, "index", i)
		ok := handler(ctx, co.q, d)
		if co.ops.autoAck {
			return
		}
		if ok {
			e := a.Ack(tag, false)
			if e != nil {
				log.WithContext(ctx).WithError(e).Error("consume one ack failed")
			}
			return
		}
		// get retry count
		var retryCount int32
		if v, o := d.Headers["x-retry-count"].(int32); o {
			retryCount = v + 1
		} else {
			retryCount = 1
		}
		// need to retry, requeue with set custom header
		if co.ops.nackRetry {
			if retryCount < co.ops.nackMaxRetryCount {
				d.Headers["x-retry-count"] = retryCount
				err = qu.ex.PublishByte(
					d.Body,
					WithPublishHeaders(d.Headers),
					WithPublishRouteKey(d.RoutingKey),
				)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("consume one republish failed")
				}
			} else {
				log.WithContext(ctx).Warn("maximum retry %d exceeded, discard data", co.ops.nackMaxRetryCount)
			}
		}
		e := a.Nack(tag, false, co.ops.nackRequeue)
		if e != nil {
			log.WithContext(ctx).WithError(e).Error("consume one nack failed")
		}
	}
	return
}

func (qu *Queue) beforeConsume(options ...func(*ConsumeOptions)) *Consume {
	var co Consume
	if qu.Error != nil {
		co.Error = qu.Error
		return &co
	}
	ops := getConsumeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	co.ops = *ops
	co.q = qu.ops.name
	co.consumer = tcr.NewConsumerFromConfig(
		&tcr.ConsumerConfig{
			Enabled:              true,
			QueueName:            qu.ops.name,
			ConsumerName:         co.ops.consumer,
			AutoAck:              co.ops.autoAck,
			Exclusive:            co.ops.exclusive,
			NoWait:               co.ops.noWait,
			Args:                 co.ops.args,
			QosCountOverride:     co.ops.qosPrefetchCount,
			SleepOnIdleInterval:  100,
			SleepOnErrorInterval: 100,
		},
		qu.ex.rb.pool,
	)
	return &co
}

func (qu *Queue) getBatch(channel *tcr.ChannelHost, queueName string, batchSize int, autoAck bool) (ds []amqp.Delivery, err error) {
	ds = make([]amqp.Delivery, 0)

	// Get A Batch of Messages
GetBatchLoop:
	for {
		// Break if we have a full batch
		if len(ds) == batchSize {
			break GetBatchLoop
		}

		var d amqp.Delivery
		var ok bool
		d, ok, err = channel.Channel.Get(queueName, autoAck)
		if err != nil {
			return
		}

		if !ok { // Break If empty
			break GetBatchLoop
		}

		ds = append(ds, d)
	}

	return
}

func (co *Consume) newContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if co.ops.autoRequestId {
		ctx = tracing.NewId(ctx)
	}
	return ctx
}
