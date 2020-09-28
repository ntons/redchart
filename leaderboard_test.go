package ranking

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestLeaderboard(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 9})
	err := LoadScripts(ctx, cli)
	if err != nil {
		t.Fatal(err)
	}

	const N = 100

	e1 := make([]*E, N) // entries to set
	e2 := make([]*E, N) // entries to incr
	e3 := make([]*E, N) // entries to sum
	for i := 0; i < N; i++ {
		e1[i] = &E{
			Id:    fmt.Sprintf("%d", i),
			Score: int64(rand.Intn(99999)),
		}
		e2[i] = &E{
			Id:    fmt.Sprintf("%d", i),
			Score: int64(rand.Intn(99999)),
		}
		e3[i] = &E{
			Id:    fmt.Sprintf("%d", i),
			Score: e1[i].Score + e2[i].Score,
		}
	}
	sort.Slice(e3, func(i, j int) bool { return e3[i].Score > e3[j].Score })

	cli.Del(ctx, "leaderboardtest:z", "leaderboardtest:h")
	x := GetLeaderboard(cli, "leaderboardtest", WithIdleExpire(time.Minute))

	if err = x.SetScore(ctx, e1...); err != nil {
		t.Log(luaSetScore)
		t.Fatal(err)
	}
	if err = x.IncrScore(ctx, e2...); err != nil {
		t.Log(luaIncScore)
		t.Fatal(err)
	}

	e4, err := x.GetRange(ctx, 0, 2*N)
	if err != nil {
		t.Log(luaGetRange)
		t.Fatal(err)
	}

	if len(e4) != len(e3) {
		t.Fatalf("unexpected len: %d", len(e4))
	}
	for i := 0; i < len(e4); i++ {
		if e3[i].Id != e4[i].Id {
			t.Fatalf("unexpected id: %s", e4[i].Id)
		}
		if e3[i].Score != e4[i].Score {
			t.Fatalf("unexpected score: %d", e4[i].Score)
		}
	}
}
