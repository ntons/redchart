package ranking

import (
	"context"

	"github.com/vmihailenco/msgpack/v4"
)

type Bubble struct {
	chart
}

func GetBubble(r RedisClient, name string, opts ...Option) Bubble {
	return Bubble{getChart(r, name, opts...)}
}

// append entries to the end of chart
func (x Bubble) Append(ctx context.Context, entries ...*Entry) (err error) {
	if len(entries) == 0 {
		return x.Touch(ctx)
	}
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, luaAppend, b2s(b)).Err()
}

// swap 2 entries by id
func (x Bubble) SwapById(ctx context.Context, id1, id2 string) (err error) {
	return x.runScript(ctx, luaSwapById, id1, id2).Err()
}

// swap 2 entries by rank
func (x Bubble) SwapByRank(ctx context.Context, rank1, rank2 int32) (err error) {
	return x.runScript(ctx, luaSwapByRank, rank1, rank2).Err()
}
