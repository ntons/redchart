package redchart

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type options struct {
	Capacity      int32  `msgpack:"capacity,omitempty"`
	NoTrim        bool   `msgpack:"no_trim"`
	ConstructFrom string `msgpack:"construct_from,omitempty"`
	NoInfo        bool   `msgpack:"no_info,omitempty"`
	ExpireAt      string `msgpack:"expire_at,omitempty"`
	Expire        string `msgpack:"expire,omitempty"`
	Persist       bool   `msgpack:"persist,omitempty"`
}

func (opts *options) encode() string {
	if b, err := msgpack.Marshal(&opts); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

type Option struct {
	apply func(o *options)
}

func applyOptions(opts []Option) string {
	_opts := &options{}
	for _, opt := range opts {
		opt.apply(_opts)
	}
	return _opts.encode()
}

// 榜单的最大容量
func WithCapacity(capacity int32) Option {
	return Option{func(o *options) { o.Capacity = capacity }}
}

// 默认情况下，在达到榜单最大容量时将淘汰末位对象
// NoTrim修改此行为为超过容量报错
func WithNoTrim() Option {
	return Option{func(o *options) { o.NoTrim = true }}
}

// 当请求榜单不存在时从指定榜单复制，可以用来制造榜单快照
func WithConstructFrom(name string) Option {
	return Option{func(o *options) { o.ConstructFrom = name }}
}

// 在进行榜单复制时忽略Info数据，只保留排名数据
// 还有另外一个作用，查询榜单时指定该参数将不返回Info
func WithNoInfo() Option {
	return Option{func(o *options) { o.NoInfo = true }}
}

// 设置过期时间，以秒为单位的Unix时间戳
// 参数为空则不对过期时间进行修改
func WithExpireAt(t time.Time) Option {
	return Option{func(o *options) {
		if t.IsZero() {
			o.ExpireAt = ""
		} else {
			o.ExpireAt = fmt.Sprintf("%d", t.Unix())
		}
	}}
}

// 设置过期时间，从当前时间算起的秒数
// 如果和ExpireAt同时设置，该参数将被忽略
// 参数为空则不对过期时间进行修改
func WithExpire(d time.Duration) Option {
	return Option{func(o *options) {
		if d <= 0 {
			o.Expire = ""
		} else {
			o.Expire = fmt.Sprintf("%d", d/time.Second)
		}
	}}
}

// 设置为不过期，不过期是默认行为
// 该参数只为清除之前设置的超时时间
func WithPersist() Option {
	return Option{func(o *options) { o.Persist = true }}
}
