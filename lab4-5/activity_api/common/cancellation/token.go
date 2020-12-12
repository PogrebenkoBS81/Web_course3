package cancellation

import (
	"context"
	"sync"
)

type Token struct {
	parent context.Context
	ctx    context.Context
	f      context.CancelFunc
	wg     sync.WaitGroup
	mtx    sync.RWMutex
	delta  int
}

// NewToken - returns new cancellation token.
func NewToken() *Token {
	t := new(Token)
	t.parent = context.Background()
	t.ctx, t.f = context.WithCancel(t.parent)
	t.delta = 1
	t.wg.Add(t.delta)

	return t
}

// NewToken - returns new custom cancellation token.
func NewCustomToken(parent context.Context, delta int) *Token {
	t := new(Token)
	t.parent = parent
	t.ctx, t.f = context.WithCancel(t.parent)
	t.delta = delta
	t.wg.Add(t.delta)

	return t
}

// Alive - returns true if context is still usable.
func (t *Token) Alive() bool {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.alive()
}

// alive - context nil error indicates usability.
func (t *Token) alive() bool {
	return t.parent.Err() == nil && t.ctx.Err() == nil
}

// Cancelled - cancellation channel.
func (t *Token) Cancelled() <-chan struct{} {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.ctx.Done()
}

// Context - returns underlying context.
func (t *Token) Context() context.Context {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.ctx
}

// Done - subtracts from 'in use' counter, when the counter reaches the zero.
func (t *Token) Done() *Token {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	t.wg.Done()

	return t
}

// Cancel - cancels underline context, requires Done() to be called explicitly.
func (t *Token) Cancel() *Token {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	t.f()

	return t
}

// Close - immediately cancels and finalizes the token, calling Done() on closed token panics.
func (t *Token) Close() *Token {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	t.f()

	for i := 0; i < t.delta; i++ {
		t.wg.Done()
	}

	return t
}

// Reset - resets the token to its initial state, no-op if token is still alive.
func (t *Token) Reset() *Token {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	if t.alive() {
		return t
	}

	t.ctx, t.f = context.WithCancel(t.parent)
	t.wg.Add(t.delta)

	return t
}

// Await - awaits until Done().
func (t *Token) Await() *Token {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	t.wg.Wait()

	return t
}
