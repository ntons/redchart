package rank

import (
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type Options struct {
	Capacity int
	// duplicate from name IF NOT EXIST
	ConstructFrom string
	// only one of ExpireAt and IdleExpire should be specified,
	// If both were set, ExpireAt was prefered
	ExpireAt   time.Time
	IdleExpire time.Duration
}

func (o *Options) encode() string {
	if o == nil {
		return ""
	}
	s := struct {
		Capacity      int    `msgpack:"capacity,omitempty"`
		ConstructFrom string `msgpack:"construct_from,omitempty"`
		ExpireAt      string `msgpack:"expire_at,omitempty"`
		IdleExpire    string `msgpack:"idle_expire,omitempty"`
	}{
		Capacity:      o.Capacity,
		ConstructFrom: o.ConstructFrom,
	}
	if !o.ExpireAt.IsZero() {
		s.ExpireAt = strconv.Itoa(int(o.ExpireAt.UnixNano() / 1e6))
	}
	if o.IdleExpire > 0 {
		s.IdleExpire = strconv.Itoa(int(o.IdleExpire / 1e6))
	}
	if b, err := msgpack.Marshal(&s); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}
