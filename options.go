package redchart

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type Options struct {
	// 榜单的最大容量
	Capacity int32 `msgpack:"capacity,omitempty"`
	// 默认情况下，在达到榜单最大容量时将淘汰末位对象
	// NoTrim修改此行为为超过容量报错
	NoTrim bool `msgpack:"no_trim,omitempty"`
	// 当请求榜单不存在时从指定榜单复制，可以用来制造榜单快照
	ConstructFrom string `msgpack:"construct_from,omitempty"`
	// 在进行榜单复制时忽略Info数据，只保留排名数据
	// 还有另外一个作用，查询榜单时指定该参数将不返回Info
	NoInfo bool `msgpack:"no_info,omitempty"`
	// 设置过期时间，以秒为单位的Unix时间戳
	// 参数为空则不对过期时间进行修改
	ExpireAt string `msgpack:"expire_at,omitempty"`
	// 设置过期时间，从当前时间算起的秒数
	// 如果和ExpireAt同时设置，该参数将被忽略
	// 参数为空则不对过期时间进行修改
	Expire string `msgpack:"expire,omitempty"`
	// 设置为不过期，不过期是默认行为
	// 该参数只为清除Expire(At)设置的超时时间
	Persist bool `msgpack:"persist,omitempty"`
	// 只对Set生效的参数
	Set SetOptions `msgpack:"set,omitempty"`
}

// 只对Set方法生效的参数
type SetOptions struct {
	// 只增加新的元素，已有元素不进行更新，OnlyAdd和OnlyUpdate不能同时为true
	// ZADD.NX
	OnlyAdd bool `msgpack:"only_add,omitempty"`
	// 只更新已有元素，不新增元素，OnlyAdd和OnlyUpdate不能同时为true
	// ZADD.XX
	OnlyUpdate bool `msgpack:"only_update,omitempty"`
	// 对分数进行增加而不是替换
	// ZADD.INCR
	IncrBy bool `msgpack:"incr_by,omitempty"`
}

func (opts Options) encode() string {
	if b, err := msgpack.Marshal(opts); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

type Option struct {
	apply func(o *Options)
}

func WithCapacity(capacity int32) Option {
	return Option{func(o *Options) { o.Capacity = capacity }}
}

func WithNoTrim(v bool) Option {
	return Option{func(o *Options) { o.NoTrim = v }}
}

func WithConstructFrom(name string) Option {
	return Option{func(o *Options) { o.ConstructFrom = name }}
}

func WithNoInfo(v bool) Option {
	return Option{func(o *Options) { o.NoInfo = v }}
}

func WithExpireAt(t time.Time) Option {
	return Option{func(o *Options) {
		if t.IsZero() {
			o.ExpireAt = ""
		} else {
			o.ExpireAt = fmt.Sprintf("%d", t.Unix())
		}
	}}
}

func WithExpire(d time.Duration) Option {
	return Option{func(o *Options) {
		if d <= 0 {
			o.Expire = ""
		} else {
			o.Expire = fmt.Sprintf("%d", d/time.Second)
		}
	}}
}

func WithPersist(v bool) Option {
	return Option{func(o *Options) { o.Persist = v }}
}

func WithSetOnlyAdd(v bool) Option {
	return Option{func(o *Options) { o.Set.OnlyAdd = v }}
}

func WithSetOnlyUpdate(v bool) Option {
	return Option{func(o *Options) { o.Set.OnlyUpdate = v }}
}

func WithSetIncrBy(v bool) Option {
	return Option{func(o *Options) { o.Set.IncrBy = v }}
}
