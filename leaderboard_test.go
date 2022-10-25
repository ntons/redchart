package redchart

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func getTestLeaderboard(ctx context.Context) Leaderboard {
	const name = "leaderboardtest"

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 9})

	cli.Del(ctx, name+":z", name+":h")

	return GetLeaderboard(cli, "leaderboardtest", WithExpire(time.Minute))
}

func TestLeaderboardSet(t *testing.T) {
	ctx := context.Background()

	lb := getTestLeaderboard(ctx)

	e := Entry{Id: "1024", Score: 1024, Info: "1024"}
	if r, err := lb.SetOne(ctx, e); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != 0 || r.Id != e.Id || r.Score != e.Score || r.Info != e.Info {
		log.Fatalf("unexpected ret: %v", r)
	}

	e.Score = e.Score + 1
	e.Info = "1025"
	if r, err := lb.SetOne(ctx, e, WithSetOnlyAdd(true)); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != 0 || r.Id != e.Id || r.Score != e.Score-1 || r.Info != "1025" {
		log.Fatalf("unexpected ret: %v", r)
	}

	e.Score = e.Score + 1
	e.Info = "1026"
	if r, err := lb.SetOne(ctx, e, WithSetOnlyUpdate(true)); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != 0 || r.Id != e.Id || r.Score != e.Score || r.Info != e.Info {
		log.Fatalf("unexpected ret: %v", r)
	}

	e.Score = e.Score + 1
	e.Info = "1027"
	if r, err := lb.SetOne(ctx, e); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != 0 || r.Id != e.Id || r.Score != e.Score || r.Info != e.Info {
		log.Fatalf("unexpected ret: %v", r)
	}

	if r, err := lb.SetOne(ctx, e, WithSetIncrBy(true)); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != 0 || r.Id != e.Id || r.Score != e.Score*2 || r.Info != e.Info {
		log.Fatalf("unexpected ret: %v", r)
	}

	e = Entry{Id: "2024", Score: 1024, Info: "1024"}
	if r, err := lb.SetOne(ctx, e, WithSetOnlyUpdate(true)); err != nil {
		log.Fatalf("failed to set: %v", err)
	} else if r.Rank != -1 || r.Id != e.Id || r.Score != 0 || r.Info != "" {
		log.Fatalf("unexpected ret: %v", r)
	}
}

func TestLeaderboardCapacity(t *testing.T) {
	ctx := context.Background()

	lb := getTestLeaderboard(ctx)

	// 没有no_trim的情况下，应该每次都能插入进去，但是总长度一直为10
	for i := int64(1); i <= 20; i++ {
		r, err := lb.Set(ctx,
			[]Entry{
				Entry{Id: fmt.Sprintf("%d", i), Score: i},
			},
			WithCapacity(10),
		)
		if err != nil {
			t.Fatalf("failed to set: %v", err)
		}
		if len(r) != 1 || r[0].Rank != 0 {
			t.Fatalf("unexpected set ret: %v", r)
		}
	}
	{
		r, err := lb.GetByRank(ctx, 0, 100)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if len(r) != 10 {
			t.Fatalf("unexpected capacity: %v", len(r))
		}
	}

	// 有no_trim情况下，超过capacity的新增将返回rank -1
	for i := int64(21); i <= 30; i++ {
		r, err := lb.Set(ctx,
			[]Entry{
				Entry{Id: fmt.Sprintf("%d", i), Score: i},
			},
			WithCapacity(10),
			WithNoTrim(true),
		)
		if err != nil {
			t.Fatalf("failed to set: %v", err)
		}
		if len(r) != 1 || r[0].Rank != -1 {
			t.Fatalf("unexpected set ret: %v", r)
		}
	}
	{
		r, err := lb.GetByRank(ctx, 0, 100)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if len(r) != 10 {
			t.Fatalf("unexpected capacity: %v", len(r))
		}
	}
}
