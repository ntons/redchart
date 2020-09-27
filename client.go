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

func (c Client) GetBubble(name string, opts ...Option) Bubble {
	return GetBubble(c.client, name, append(c.opts, opts...)...)
}
