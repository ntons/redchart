package redchart

import (
	"context"

	"github.com/ntons/redis"
	"github.com/vmihailenco/msgpack/v4"
)

type BubbleChart struct {
	chart
}

func GetBubbleChart(rcli redis.Client, name string, opts ...Option) BubbleChart {
	return BubbleChart{getChart(rcli, name, opts)}
}

// append entries to the end of chart
func (x BubbleChart) Append(
	ctx context.Context, entries []Entry, opts ...Option) (err error) {
	if len(entries) == 0 {
		return x.Touch(ctx)
	}
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, opts, luaAppend, b2s(b)).Err()
}

// swap 2 entries by id
func (x BubbleChart) SwapById(
	ctx context.Context, id1, id2 string, opts ...Option) (err error) {
	return x.runScript(ctx, opts, luaSwapById, id1, id2).Err()
}

// swap 2 entries by rank
func (x BubbleChart) SwapByRank(
	ctx context.Context, rank1, rank2 int32, opts ...Option) (err error) {
	return x.runScript(ctx, opts, luaSwapByRank, rank1, rank2).Err()
}
