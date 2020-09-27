package ranking

import (
	"context"
	"fmt"

	"github.com/ntons/tongo/tunsafe"
	"github.com/vmihailenco/msgpack/v4"
)

type Leaderboard struct {
	Base
}

func GetLeaderboard(r Client, name string, opts *Options) Leaderboard {
	return Leaderboard{newBase(r, name, opts)}
}

// update rank by score, e.Rank will be ignored
func (l Leaderboard) SetScore(
	ctx context.Context, es ...*Entry) (err error) {
	b, err := msgpack.Marshal(es)
	if err != nil {
		return
	}
	return l.eval(ctx, luaSetScore, tunsafe.BytesToString(b)).Err()
}

// update score by increment, e.Rank will be ignored
// e.Score will be set to updated value if success
func (l Leaderboard) IncScore(
	ctx context.Context, es ...*Entry) (err error) {
	b, err := msgpack.Marshal(es)
	if err != nil {
		return
	}
	s, err := l.eval(ctx, luaIncScore, tunsafe.BytesToString(b)).Text()
	if err != nil {
		return
	}
	var r []int64
	if err = msgpack.Unmarshal(tunsafe.StringToBytes(s), &r); err != nil {
		return
	}
	if len(r) != len(es) {
		return fmt.Errorf("invalid return size: %d != %d", len(r), len(es))
	}
	for i, v := range r {
		es[i].Score = int64(v)
	}
	return
}
