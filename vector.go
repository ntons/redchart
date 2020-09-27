package ranking

import (
	"context"

	"github.com/ntons/tongo/tunsafe"
	"github.com/vmihailenco/msgpack/v4"
)

type Vector struct {
	Base
}

func GetVector(r Client, name string, opts *Options) Vector {
	return Vector{newBase(r, name, opts)}
}

func (v Vector) Append(ctx context.Context, es ...*Entry) (err error) {
	b, err := msgpack.Marshal(es)
	if err != nil {
		return
	}
	return v.eval(ctx, luaAppend, tunsafe.BytesToString(b)).Err()
}

func (v Vector) SwapById(ctx context.Context, id1, id2 string) (err error) {
	return v.eval(ctx, luaSwapById, id1, id2).Err()
}

func (v Vector) SwapByRank(ctx context.Context, rank1, rank2 int) (err error) {
	return v.eval(ctx, luaSwapByRank, rank1, rank2).Err()
}
