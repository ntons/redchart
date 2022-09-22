package ranking

import (
	"context"
	"fmt"

	"github.com/vmihailenco/msgpack/v4"
)

type Leaderboard struct {
	chart
}

func GetLeaderboard(r RedisClient, name string, opts ...Option) Leaderboard {
	return Leaderboard{getChart(r, name, opts...)}
}

// add entries which not exist to chart
func (x Leaderboard) Add(
	ctx context.Context, entries ...*Entry) (err error) {
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, luaAdd, b2s(b)).Err()
}

// update score or add, e.Rank will be ignored
func (x Leaderboard) Set(
	ctx context.Context, entries ...*Entry) (err error) {
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, luaSet, b2s(b)).Err()
}

// update score by increment, e.Rank will be ignored
// e.Score will be set to updated value if success
func (x Leaderboard) Incr(
	ctx context.Context, entries ...*Entry) (err error) {
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	s, err := x.runScript(ctx, luaIncr, b2s(b)).Text()
	if err != nil {
		return
	}
	var r []int64
	if err = msgpack.Unmarshal(s2b(s), &r); err != nil {
		return
	}
	if len(r) != len(entries) {
		return fmt.Errorf("invalid return size: %d != %d", len(r), len(entries))
	}
	for i, v := range r {
		entries[i].Score = int64(v)
	}
	return
}
