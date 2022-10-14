package redchart

import (
	"context"

	"github.com/ntons/redis"
	"github.com/vmihailenco/msgpack/v4"
)

type Leaderboard struct {
	chart
}

func GetLeaderboard(rcli redis.Client, name string, opts ...Option) Leaderboard {
	return Leaderboard{getChart(rcli, name, opts)}
}

func (x Leaderboard) Set(
	ctx context.Context, entries []Entry, opts ...Option) (_ []Entry, err error) {
	buf, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	str, err := x.runScript(ctx, opts, luaSet, b2s(buf)).Text()
	if err != nil {
		return
	}
	var ret []Entry
	if err = msgpack.Unmarshal(s2b(str), &ret); err != nil {
		return
	}
	return ret, nil
}

func (x Leaderboard) SetOne(
	ctx context.Context, entry Entry, opts ...Option) (_ Entry, err error) {
	ret, err := x.Set(ctx, []Entry{entry}, opts...)
	if err != nil {
		return
	}
	return ret[0], nil
}

type RandByScoreArg struct {
	Min   int64 `json:"min" msgpack:"min"`
	Max   int64 `json:"max" msgpack:"max"`
	Count int   `json:"count" msgpack:"count"`
}

// get entries randomly by score
func (x Leaderboard) RandByScore(
	ctx context.Context, args RandByScoreArg, opts ...Option) (entries []Entry, err error) {
	b, err := msgpack.Marshal(args)
	if err != nil {
		return
	}
	s, err := x.runScript(ctx, opts, luaRandByScore, b2s(b)).Text()
	if err != nil {
		return
	}
	var ids []string
	if err = msgpack.Unmarshal(s2b(s), &ids); err != nil {
		return
	}
	return x.GetById(ctx, ids, opts...)
}
