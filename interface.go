package idempotentcache

import "context"

type Status int

//go:generate mockgen -source=api.go -destination=./mocks/mocks.go -package=mocks

const (
	StatusUnknown Status = iota
	StatusFirstSeen
	StatusAwaitingResponse
	StatusHandled
)

type Interface interface {
	Handled(ctx context.Context, requestId string, resp []byte) error
	Invalidate(ctx context.Context, requestId string) error
	Receive(ctx context.Context, requestId string, respOut *[]byte) (Status, error)
	ReceiveAndWait(ctx context.Context, requestId string, respOut *[]byte) (Status, error)
	Wait(ctx context.Context, requestId string, respOut *[]byte) error
}
