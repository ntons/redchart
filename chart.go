package redchart

import (
	"context"

	"github.com/ntons/redis"
	"github.com/vmihailenco/msgpack/v4"
)

type chart struct {
	rcli redis.Client
	name string
	opts []Option
}

func getChart(rcli redis.Client, name string, opts []Option) chart {
	return chart{rcli: rcli, name: name, opts: opts}
}

// Touch modify metadata over chart options
func (x chart) Touch(ctx context.Context, opts ...Option) (err error) {
	if err = x.runScript(ctx, opts, luaTouch).Err(); err == redis.Nil {
		err = nil
	}
	return
}

// get entries by range
func (x chart) GetByRank(
	ctx context.Context, offset, count int32, opts ...Option) (entries []Entry, err error) {
	s, err := x.runScript(ctx, opts, luaGetByRank, offset, offset+count-1).Text()
	if err != nil {
		return
	}
	if err = msgpack.Unmarshal(s2b(s), &entries); err != nil {
		return
	}
	return
}

func (x chart) GetById(
	ctx context.Context, ids []string, opts ...Option) (entries []Entry, err error) {
	s, err := x.runScriptString(ctx, opts, luaGetById, ids...).Text()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	if err = msgpack.Unmarshal(s2b(s), &entries); err != nil {
		return
	}
	return
}

// remove chart entry by id
func (x chart) RemoveById(
	ctx context.Context, ids []string, opts ...Option) (err error) {
	if len(ids) == 0 {
		return x.Touch(ctx)
	}
	return x.runScriptString(ctx, opts, luaRemoveId, ids...).Err()
}

// set entry info
func (x chart) SetInfo(
	ctx context.Context, entries []Entry, opts ...Option) (err error) {
	if len(entries) == 0 {
		return x.Touch(ctx)
	}
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, opts, luaSetInfo, b2s(b)).Err()
}

func (x chart) runScript(
	ctx context.Context, opts []Option, script *redis.Script, args ...interface{}) *redis.Cmd {
	var o Options
	for _, opt := range x.opts {
		opt.apply(&o)
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	args = append([]interface{}{o.encode()}, args...)
	return script.Run(ctx, x.rcli, []string{x.name}, args...)
}

func (x chart) runScriptString(
	ctx context.Context, opts []Option, script *redis.Script, args ...string) *redis.Cmd {
	tmp := make([]interface{}, 0, len(args))
	for _, arg := range args {
		tmp = append(tmp, arg)
	}
	return x.runScript(ctx, opts, script, tmp...)
}
