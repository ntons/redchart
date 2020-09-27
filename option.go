package ranking

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type options struct {
	// capacity of chart
	Capacity int32 `msgpack:"capacity,omitempty"`
	// duplicate from name IF NOT EXIST
	ConstructFrom string `msgpack:"construct_from,omitempty"`
	// only one of ExpireAt and IdleExpire should be specified,
	// If both were set, ExpireAt was prefered
	ExpireAt   string `msgpack:"expire_at,omitempty"`
	IdleExpire string `msgpack:"idle_expire,omitempty"`
}

func (o *options) encode() string {
	if o == nil {
		return ""
	}
	if b, err := msgpack.Marshal(o); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

type Option interface {
	apply(o *options)
}

type funcOption struct {
	fn func(o *options)
}

func (f funcOption) apply(o *options) {
	f.fn(o)
}

func WithCapacity(capacity int32) Option {
	return funcOption{func(o *options) { o.Capacity = capacity }}
}
func ConstructFrom(name string) Option {
	return funcOption{func(o *options) { o.ConstructFrom = name }}
}
func ExpireAt(t time.Time) Option {
	return funcOption{func(o *options) {
		if t.IsZero() {
			o.ExpireAt = ""
		} else {
			o.ExpireAt = fmt.Sprintf("%d", t.UnixNano()/1e6)
		}
	}}
}
func IdleExpire(d time.Duration) Option {
	return funcOption{func(o *options) {
		if d == 0 {
			o.IdleExpire = ""
		} else {
			o.IdleExpire = fmt.Sprintf("%d", d/time.Millisecond)
		}
	}}
}
