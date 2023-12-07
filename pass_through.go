package apicache

import "context"

// PassThroughClient is a client that behaves like there is no cache and returns
// sensible values for all operations so the caller can proceed normally; however,
// it will not handle idempotent use-cases correctly.
type PassThroughClient struct {}

var _ Interface = (*PassThroughClient)(nil)

func (c *PassThroughClient) Handled(ctx context.Context, requestId string, resp []byte) error {
	return nil
}

func (c *PassThroughClient) Invalidate(ctx context.Context, requestId string) error {
	return nil
}

func (c *PassThroughClient) Receive(ctx context.Context, requestId string, respOut *[]byte) (Status, error) {
	return StatusFirstSeen, nil
}

func (c *PassThroughClient) ReceiveAndWait(ctx context.Context, requestId string, respOut *[]byte) (Status, error) {
	return StatusFirstSeen, nil
}

func (c *PassThroughClient) Wait(ctx context.Context, requestId string, respOut *[]byte) error {
	return nil
}
