package ranking

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type options struct {
	// capacity of chart
	Capacity int32 `msgpack:"capacity,omitempty"`
	// do not trim if exceed capacity
	NotTrim bool `msgpack:"not_trim"`
	// duplicate from name IF NOT EXIST
	ConstructFrom string `msgpack:"construct_from,omitempty"`
	// not expire is the default behavious,
	// this is used to unset blowing expiration options
	NotExpire bool `msgpack:"not_expire,omitempty"`
	// only one of ExpireAt and IdleExpire should be specified,
	// If both were set, ExpireAt was prefered
	ExpireAt   string `msgpack:"expire_at,omitempty"`
	IdleExpire string `msgpack:"idle_expire,omitempty"`
	// query chart with no info return(only rank/id/score)
	NoInfo bool `msgpack:"no_info,omitempty"`
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
func WithNotTrim() Option {
	return funcOption{func(o *options) { o.NotTrim = true }}
}
func WithConstructFrom(name string) Option {
	return funcOption{func(o *options) { o.ConstructFrom = name }}
}
func WithNotExpire() Option {
	return funcOption{func(o *options) {
		o.NotExpire = true
	}}
}
func WithExpireAt(t time.Time) Option {
	return funcOption{func(o *options) {
		if t.IsZero() {
			o.ExpireAt = ""
		} else {
			o.ExpireAt = fmt.Sprintf("%d", t.UnixNano()/1e6)
		}
	}}
}
func WithIdleExpire(d time.Duration) Option {
	return funcOption{func(o *options) {
		if d <= 0 {
			o.IdleExpire = ""
		} else {
			o.IdleExpire = fmt.Sprintf("%d", d/time.Millisecond)
		}
	}}
}
func WithNoInfo() Option {
	return funcOption{func(o *options) { o.NoInfo = true }}
}
