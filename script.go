package ranking

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/go-redis/redis/v8"
)

type scripter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
}

type ScriptOption interface {
	apply(*Script)
}

type funcScriptOption struct {
	fn func(*Script)
}

func (o funcScriptOption) apply(s *Script) {
	o.fn(s)
}

func WithTryShaProb(prob int) ScriptOption {
	return funcScriptOption{func(s *Script) { s.tryShaProb = prob }}
}

// scripts create by remon.NewScript
var scripts []*Script

// load all scripts to redis
// NewScript must be invoked before LoadScripts
func LoadScripts(ctx context.Context, c scripter) (err error) {
	for _, s := range scripts {
		if _, err = s.Load(ctx, c).Result(); err != nil {
			return
		}
	}
	return
}

// Script is similar to redis.Script, but optimize evalsha strategy
type Script struct {
	src        string
	hash       string
	loaded     bool // is script loaded
	tryShaProb int  // trySha probability 1/n
}

func newScript(src string, opts ...ScriptOption) *Script {
	src = fmt.Sprintf(luaTemplate, src)
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	s := &Script{
		src:        src,
		hash:       hex.EncodeToString(h.Sum(nil)),
		loaded:     true,
		tryShaProb: 100, // 1% by default
	}
	for _, opt := range opts {
		opt.apply(s)
	}
	scripts = append(scripts, s)
	return s
}

func (s *Script) Hash() string {
	return s.hash
}

func (s *Script) Load(ctx context.Context, c scripter) *redis.StringCmd {
	return c.ScriptLoad(ctx, s.src)
}

func (s *Script) Exists(ctx context.Context, c scripter) *redis.BoolSliceCmd {
	return c.ScriptExists(ctx, s.hash)
}

func (s *Script) Eval(ctx context.Context, c scripter, keys []string, args ...interface{}) *redis.Cmd {
	return c.Eval(ctx, s.src, keys, args...)
}

func (s *Script) EvalSha(ctx context.Context, c scripter, keys []string, args ...interface{}) *redis.Cmd {
	return c.EvalSha(ctx, s.hash, keys, args...)
}

func (s *Script) Run(ctx context.Context, c scripter, keys []string, args ...interface{}) *redis.Cmd {
	if !s.loaded {
		s.loaded = s.trySha()
	}
	if s.loaded {
		r := s.EvalSha(ctx, c, keys, args...)
		if err := r.Err(); err == nil || !strings.HasPrefix(err.Error(), "NOSCRIPT ") {
			return r
		}
		s.loaded = false
	}
	return s.Eval(ctx, c, keys, args...)
}

func (s *Script) trySha() bool {
	return rand.Intn(100) == 0
}