package port

import (
	"context"
	"testing"

	mtfctx "github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/match"
)

type Kind int

const (
	KIND_SERVER        Kind = 1
	KIND_CLIENT             = iota
	KIND_MESSAGE_QEUEU      = iota
)

type PortImpl interface {
	Send(ctx context.Context, msg interface{}) error
	Receive(ctx context.Context) (interface{}, error)
	Kind() Kind
	Name() string
}

type Port struct {
	impl PortImpl
}

type sendOptions struct {
	ctx context.Context
}

type SendOption func(*sendOptions)

func WithCtx(ctx context.Context) SendOption {
	return func(o *sendOptions) {
		o.ctx = ctx
	}
}

func (p *Port) Send(t *testing.T, i interface{}, opts ...SendOption) error {
	defOpts := &sendOptions{
		ctx: context.Background(),
	}

	for _, o := range opts {
		o(defOpts)
	}

	if err := p.impl.Send(defOpts.ctx, i); err != nil {
		t.Fatalf("failed to send %T from %s, err: %v", i, p.impl.Name(), err)
	}
	mtfc := mtfctx.Get(t)
	mtfc.LogSend(p.impl.Name(), i)
	return nil
}

func (p *Port) Receive(t *testing.T, i interface{}) error {
	ctx := context.Background()
	m, err := p.impl.Receive(ctx)
	mtfc := mtfctx.Get(t)
	mtfc.LogReceive(p.impl.Name(), m)
	if err != nil {
		t.Fatalf("failed to receive %T from %s: %v", i, p.impl.Name(), err)
	}

	switch t := i.(type) {
	case match.FnMatcher:
		t.Match(err, m)
	case match.Any:
	default:
	}

	return nil
}
