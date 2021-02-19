package ranking

type Client struct {
	client RedisClient
	opts   []Option
}

func New(client RedisClient, opts ...Option) Client {
	return Client{client: client, opts: opts}
}

func (c Client) GetLeaderboard(name string, opts ...Option) Leaderboard {
	return GetLeaderboard(c.client, name, append(c.opts, opts...)...)
}

func (c Client) GetBubbleChart(name string, opts ...Option) BubbleChart {
	return GetBubbleChart(c.client, name, append(c.opts, opts...)...)
}
