package hub

import (
	"context"
	"sync/atomic"
)

type MessageChannel[T any] struct {
	channel    chan T
	context    context.Context
	bufferSize int
	closed     atomic.Int32
}

func NewMessageChannel[T any](ctx context.Context, bufferSize int) *MessageChannel[T] {
	return &MessageChannel[T]{
		channel:    make(chan T, bufferSize),
		context:    ctx,
		bufferSize: bufferSize,
	}
}

func (mc *MessageChannel[T]) Send(ctx context.Context, message T) error {
	select {
	case mc.channel <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-mc.context.Done():
		return mc.context.Err()
	}
}

func (mc *MessageChannel[T]) Receive(ctx context.Context) (T, error) {
	select {
	case message := <-mc.channel:
		return message, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-mc.context.Done():
		var zero T
		return zero, mc.context.Err()
	}
}

func (mc *MessageChannel[T]) TryReceive() (T, bool) {
	select {
	case message := <-mc.channel:
		return message, true
	default:
		var zero T
		return zero, false
	}
}

func (mc *MessageChannel[T]) Close() {
	if mc.closed.CompareAndSwap(0, 1) {
		close(mc.channel)
	}
}

func (mc *MessageChannel[T]) IsClosed() bool {
	return mc.closed.Load() == 1
}

func (mc *MessageChannel[T]) BufferSize() int {
	return mc.bufferSize
}

func (mc *MessageChannel[T]) QueueLength() int {
	return len(mc.channel)
}
