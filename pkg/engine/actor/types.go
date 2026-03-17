// Package actor provides the actor framework for engine channel refactoring.
//
// This package implements the actor pattern for managing shared state in the
// engine package. Each actor owns a specific state domain and processes
// messages sequentially via a channel, eliminating lock contention.
//
// The framework provides:
//   - Actor interface for implementing stateful actors
//   - BaseMessage for standard request-reply patterns
//   - Coordinator for managing actor lifecycles
//   - Helper functions for synchronous message passing
package actor

import (
	"context"
	"time"
)

// Actor is the interface for all actors in the system.
// An actor processes messages sequentially from its inbox channel,
// ensuring exclusive access to its internal state without locks.
type Actor interface {
	// Run starts the actor's event loop. It should block until ctx is cancelled.
	// The actor should drain its inbox channel and process each message
	// before returning.
	Run(ctx context.Context)
}

// Reply is the standard response type for actor messages.
// It carries either a result or an error, similar to function returns.
type Reply struct {
	Result any
	Error  error
}

// BaseMessage provides common message functionality for request-reply patterns.
// Each message carries a reply channel for synchronous communication.
type BaseMessage struct {
	replyCh chan Reply
}

// NewBaseMessage creates a new base message with a reply channel.
// It returns the message and a receive-only channel for the caller to wait on.
func NewBaseMessage() (BaseMessage, <-chan Reply) {
	ch := make(chan Reply, 1)
	return BaseMessage{replyCh: ch}, ch
}

// Reply returns the reply channel for the message.
// The actor should send exactly one reply on this channel.
func (m BaseMessage) Reply() chan<- Reply {
	return m.replyCh
}

// SendAndWait sends a message to an inbox and waits for a reply with timeout.
// This is a helper for synchronous-style communication with actors.
//
// Type parameter T is the message type (must match the inbox channel type).
//
// Example:
//
//	msg, replyCh := actor.NewBaseMessage()
//	getMsg := &LRUGetMsg{BaseMessage: msg, ID: id}
//	result, err := actor.SendAndWait(cache.inbox, getMsg, replyCh, 5*time.Second)
func SendAndWait[T any](inbox chan<- T, msg T, replyCh <-chan Reply, timeout time.Duration) (any, error) {
	select {
	case inbox <- msg:
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}

	select {
	case reply := <-replyCh:
		return reply.Result, reply.Error
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

// SendNoWait sends a message without waiting for a reply.
// This is used for fire-and-forget messages like notifications.
func SendNoWait[T any](inbox chan<- T, msg T) error {
	select {
	case inbox <- msg:
		return nil
	default:
		return context.DeadlineExceeded
	}
}

// DefaultTimeout is the default timeout for actor messages.
const DefaultTimeout = 5 * time.Second

// Send sends a message to an inbox with timeout.
// Returns an error if the send times out before the message is queued.
func Send[T any](inbox chan<- T, msg T, timeout time.Duration) error {
	select {
	case inbox <- msg:
		return nil
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// CallResult extracts a typed result from a reply channel with timeout.
// This helper eliminates repetitive type assertion code in compatibility wrappers.
//
// Example:
//
//	msg, replyCh := actor.NewBaseMessage()
//	inbox <- &lruLenMsg{BaseMessage: msg}
//	count, err := actor.CallResult[int](replyCh, timeout)
func CallResult[R any](replyCh <-chan Reply, timeout time.Duration) (R, error) {
	var zero R
	select {
	case reply := <-replyCh:
		if reply.Error != nil {
			return zero, reply.Error
		}
		if reply.Result == nil {
			return zero, nil
		}
		return reply.Result.(R), nil
	case <-time.After(timeout):
		return zero, context.DeadlineExceeded
	}
}
