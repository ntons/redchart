package ranking

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestBubble(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	err := LoadScripts(ctx, cli)
	if err != nil {
		t.Fatal(err)
	}

	const N = 100

	e1 := make([]*E, N)
	for i := 0; i < N; i++ {
		e1[i] = &E{Id: fmt.Sprintf("%d", i)}
	}

	cli.Del(ctx, "bubbletest:z", "bubbletest:h")
	x := GetBubble(cli, "bubbletest", WithIdleExpire(time.Minute))

	if err = x.Append(ctx, e1...); err != nil {
		t.Fatal(err)
	}
	e2, err := x.GetRange(ctx, 0, N)
	if err != nil {
		t.Fatal(err)
	}
	if len(e2) != N {
		t.Fatalf("unexpected len: %d", len(e2))
	}
	for i := 0; i < len(e2); i++ {
		if e1[i].Id != e2[i].Id {
			t.Fatalf("unexpected id: %s", e2[i].Id)
		}
	}

	for i := 0; i < N; i++ {
		u := rand.Int31n(N)
		v := rand.Int31n(N)
		if rand.Int31n(2) == 0 {
			x.SwapById(ctx, e1[u].Id, e1[v].Id)
		} else {
			x.SwapByRank(ctx, u, v)
		}
		e1[u], e1[v] = e1[v], e1[u]
	}

	e3, err := x.GetRange(ctx, 0, N)
	if err != nil {
		t.Fatal(err)
	}

	if len(e3) != N {
		t.Fatalf("unexpected len: %d", len(e3))
	}
	for i := 0; i < len(e3); i++ {
		if e1[i].Id != e3[i].Id {
			t.Fatalf("unexpected id: %s", e3[i].Id)
		}
	}

}
