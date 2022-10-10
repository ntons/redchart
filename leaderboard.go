package redchart

import (
	"context"
	"fmt"

	"github.com/ntons/redis"
	"github.com/vmihailenco/msgpack/v4"
)

type Leaderboard struct {
	chart
}

func GetLeaderboard(rcli redis.Client, name string, opts ...Option) Leaderboard {
	return Leaderboard{getChart(rcli, name, applyOptions(opts))}
}

func (x Leaderboard) handleRet(ret int) (err error) {
	if ret < 0 {
		switch ret {
		case -1:
			err = ErrChartFull
		default:
			err = fmt.Errorf("error code: %d", ret)
		}
	}
	return
}

// add entries which not exist to chart
func (x Leaderboard) Add(
	ctx context.Context, entries ...*Entry) (err error) {
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	ret, err := x.runScript(ctx, luaAdd, b2s(b)).Int()
	if err != nil {
		return
	}
	return x.handleRet(ret)
}

// update score or add, e.Rank will be ignored
func (x Leaderboard) Set(
	ctx context.Context, entries ...*Entry) (err error) {
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	ret, err := x.runScript(ctx, luaSet, b2s(b)).Int()
	if err != nil {
		return
	}
	return x.handleRet(ret)
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

type RandByScoreArg struct {
	Min   int64 `json:"min" msgpack:"min"`
	Max   int64 `json:"max" msgpack:"max"`
	Count int   `json:"count" msgpack:"count"`
}

// get entries randomly by score
func (x Leaderboard) RandByScore(
	ctx context.Context, args ...RandByScoreArg) (entries []*Entry, err error) {
	b, err := msgpack.Marshal(args)
	if err != nil {
		return
	}
	s, err := x.runScript(ctx, luaRandByScore, b2s(b)).Text()
	if err != nil {
		return
	}
	var ids []string
	if err = msgpack.Unmarshal(s2b(s), &ids); err != nil {
		return
	}
	return x.GetById(ctx, ids...)
}
