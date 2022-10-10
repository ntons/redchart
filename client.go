package redchart

import (
	"github.com/ntons/redis"
)

type Client struct {
	rcli redis.Client
}

func New(rcli redis.Client) Client {
	return Client{rcli: rcli}
}

func (cli Client) GetLeaderboard(name string, opts ...Option) Leaderboard {
	return GetLeaderboard(cli.rcli, name, opts...)
}

func (cli Client) GetBubbleChart(name string, opts ...Option) BubbleChart {
	return GetBubbleChart(cli.rcli, name, opts...)
}
