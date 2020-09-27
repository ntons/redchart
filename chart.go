package ranking

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v4"
)

type RedisClient interface {
	scripter
}

type Entry struct {
	Rank  int32  `json:"rank" msgpack:"rank"`
	Id    string `json:"id" msgpack:"id"`
	Info  string `json:"info" msgpack:"info"`
	Score int64  `json:"score" msgpack:"score"`
}

type E = Entry

type chart struct {
	client RedisClient
	name   string
	opts   string
}

func getChart(client RedisClient, name string, opts ...Option) chart {
	var o options
	for _, opt := range opts {
		opt.apply(&o)
	}
	return chart{client: client, name: name, opts: o.encode()}
}

// Touch modify metadata over chart options
func (x chart) Touch(ctx context.Context) (err error) {
	return x.runScript(ctx, luaTouch).Err()
}

// get entries by range
func (x chart) GetRange(
	ctx context.Context, offset, count int32) (entries []*Entry, err error) {
	s, err := x.runScript(ctx, luaGetRange, offset, offset+count-1).Text()
	if err != nil {
		return
	}
	if err = msgpack.Unmarshal(s2b(s), &entries); err != nil {
		return
	}
	return
}

// get entry by id
func (x chart) GetById(ctx context.Context, id string) (e *Entry, err error) {
	s, err := x.runScript(ctx, luaGetById, id).Text()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	e = &Entry{}
	if err = msgpack.Unmarshal(s2b(s), e); err != nil {
		return
	}
	return
}

// remove chart entry by id
func (x chart) RemoveById(ctx context.Context, ids ...string) (err error) {
	if len(ids) == 0 {
		return x.Touch(ctx)
	}
	return x.runScriptString(ctx, luaRemoveId, ids...).Err()
}

// set entry info
func (x chart) SetInfo(ctx context.Context, entries ...*Entry) (err error) {
	if len(entries) == 0 {
		return x.Touch(ctx)
	}
	b, err := msgpack.Marshal(entries)
	if err != nil {
		return
	}
	return x.runScript(ctx, luaSetInfo, b2s(b)).Err()
}

// i don't think you should remove entry by range

func (x chart) runScript(
	ctx context.Context, script *Script, args ...interface{}) *redis.Cmd {
	args = append([]interface{}{x.opts}, args...)
	return script.Run(ctx, x.client, []string{x.name}, args...)
}

func (x chart) runScriptString(
	ctx context.Context, script *Script, args ...string) *redis.Cmd {
	tmp := make([]interface{}, 0, len(args))
	for _, arg := range args {
		tmp = append(tmp, arg)
	}
	return x.runScript(ctx, script, tmp...)
}
