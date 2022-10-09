package redchart

import (
	"github.com/ntons/redis"
)

type Client struct {
	client redis.Client
	opts   []Option
}

func New(client redis.Client, opts ...Option) Client {
	return Client{client: client, opts: opts}
}

func (c Client) GetLeaderboard(name string, opts ...Option) Leaderboard {
	return GetLeaderboard(c.client, name, append(c.opts, opts...)...)
}

func (c Client) GetBubbleChart(name string, opts ...Option) BubbleChart {
	return GetBubbleChart(c.client, name, append(c.opts, opts...)...)
}
