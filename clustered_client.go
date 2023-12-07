package idempotentcache

import (
	"context"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type ClusteredClient struct {
	// the maximum amount of time to wait for a handled request. Note that
	// by leverage a context cancel or timeout, this value can be decreased
	// on a per-call basis.
	MaxWaitTimeout time.Duration

	// the time to wait between polling for a handled request
	PollInterval time.Duration

	// the redis client to use
	RedisClient *redis.ClusterClient

	// the maximum amount of time to hold the Request ID in memory. In other words,
	// the time window in which to deduplicate requests.
	Ttl time.Duration

	// the maximum amount of time to wait Handled() to be called
	// before returning a time out error
	WaitFor time.Duration
}

var _ Interface = (*ClusteredClient)(nil)

func (c *ClusteredClient) Handled(ctx context.Context, requestId string, resp []byte) error {
	if err := c.RedisClient.Set(ctx, requestId, resp, c.Ttl).Err(); err != nil {
		return fmt.Errorf("Put: redis set error: %w", err)
	}
	return nil
}

func (c *ClusteredClient) Invalidate(ctx context.Context, requestId string) error {
	result := c.RedisClient.Expire(ctx, requestId, 0)
	if result.Err() != nil {
		return fmt.Errorf("Delete: %w", result.Err())
	}
	return nil
}

func (c *ClusteredClient) Receive(ctx context.Context, requestId string, respOut *[]byte) (Status, error) {
	// TODO
	entry, err := c.RedisClient.Get(ctx, requestId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return StatusFirstSeen, nil
		}
		return StatusUnknown, fmt.Errorf("Receive: RedisClient get error: %w", err)
	}

	respOut = &entry

	return StatusUnknown, errors.New("not implemented")
}

func (c *ClusteredClient) ReceiveAndWait(ctx context.Context, requestId string, respOut *[]byte) (Status, error) {
	status, err := c.Receive(ctx, requestId, respOut)
	if err != nil {
		return status, fmt.Errorf("ReceiveAndWait: %w", err)
	}

	switch status {
	case StatusFirstSeen | StatusHandled:
		return status, nil

	case StatusAwaitingResponse:
		if err := c.Wait(ctx, requestId, respOut); err != nil {
			return status, fmt.Errorf("ReceiveAndWait: %w", err)
		}
		return StatusHandled, nil

	default:
		return status, fmt.Errorf("ReceiveAndWait: unknown status")
	}
}

func (c *ClusteredClient) Wait(ctx context.Context, requestId string, respOut *[]byte) error {
	done := make(chan bool)
	time.AfterFunc(c.MaxWaitTimeout, func() {
		done <- true
	})

	ticker := time.NewTicker(c.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-done:
			return fmt.Errorf("Wait: timed out waiting for response")

		case <-ticker.C:
			status, err := c.Receive(ctx, requestId, respOut)
			if err != nil {
				return fmt.Errorf("Wait: %w", err)
			}
			switch status {
			case StatusHandled:
				return nil

			case StatusFirstSeen:
				return fmt.Errorf("Wait: request first seen so should not be waiting")

			case StatusAwaitingResponse:
				continue

			default:
				return fmt.Errorf("Wait: unknown status")
			}
		}
	}
}
