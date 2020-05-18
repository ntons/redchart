package rank

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/ntons/tongo/tunsafe"
	"github.com/vmihailenco/msgpack/v4"
)

type Base struct {
	r    Client
	name string
	opts string
}

func newBase(r Client, name string, opts *Options) Base {
	return Base{r: r, name: name, opts: opts.encode()}
}

func (x Base) eval(ctx context.Context,
	lua *redis.Script, args ...interface{}) (cmd *redis.Cmd) {
	cmdArgs := make([]interface{}, 5, 5+len(args))
	cmdArgs[0] = "evalsha"
	cmdArgs[1] = lua.Hash()
	cmdArgs[2] = 1
	cmdArgs[3] = x.name
	cmdArgs[4] = x.opts
	cmdArgs = append(cmdArgs, args...)
	cmd = redis.NewCmd(cmdArgs...)
	x.r.ProcessContext(ctx, cmd)
	return
}

func (x Base) evalStr(
	ctx context.Context, lua *redis.Script, strs ...string) (cmd *redis.Cmd) {
	var args = make([]interface{}, 0, len(strs))
	for _, s := range strs {
		args = append(args, s)
	}
	return x.eval(ctx, lua, args...)
}

func (x Base) Client() Client { return x.r }

func (x Base) Name() string { return x.name }

func (x Base) Touch(ctx context.Context) (err error) {
	return x.eval(ctx, luaTouch).Err()
}

func (x Base) RemoveId(ctx context.Context, ids ...string) (err error) {
	return x.evalStr(ctx, luaRemoveId, ids...).Err()
}

func (x Base) SetInfo(
	ctx context.Context, es ...*Entry) (err error) {
	b, err := msgpack.Marshal(es)
	if err != nil {
		return
	}
	return x.eval(ctx, luaSetInfo, tunsafe.BytesToString(b)).Err()
}

func (x Base) GetRange(
	ctx context.Context, begin, count int) (es []*Entry, err error) {
	s, err := x.eval(ctx, luaGetRange, begin, begin+count-1).Text()
	if err != nil {
		return
	}
	if err = msgpack.Unmarshal(tunsafe.StringToBytes(s), &es); err != nil {
		return
	}
	return
}

func (x Base) GetById(
	ctx context.Context, id string) (e *Entry, err error) {
	s, err := x.eval(ctx, luaGetById, id).Text()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	e = &Entry{}
	if err = msgpack.Unmarshal(tunsafe.StringToBytes(s), e); err != nil {
		return
	}
	return
}
