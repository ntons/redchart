package redchart

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func getTestLeaderboard(ctx context.Context, opts ...Option) Leaderboard {
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

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 9})

	cli.Del(ctx, "*")

	lb := GetLeaderboard(
		cli,
		"leaderboardtest",
		WithExpire(time.Minute),
		WithCapacity(10),
		//WithNoInfo(),
		//WithNotTrim(),
	)

	elist := make([]Entry, 0)
	for i := int64(1); i <= 100; i++ {
		s := fmt.Sprintf("%d", i)
		elist = append(elist, Entry{Id: s, Info: s, Score: i})
	}

	{
		r, err := lb.Set(ctx, elist, WithSetOnlyAdd(true))
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}

	{
		r, err := lb.GetByRank(ctx, 0, 3)
		if err != nil {
			fmt.Printf("failed to rand: %v\n", err)
			return
		}
		fmt.Println(r)
	}
	{
		r, err := lb.GetByRank(ctx, 3, 6)
		if err != nil {
			fmt.Printf("failed to rand: %v\n", err)
			return
		}
		fmt.Println(r)
	}

	{
		r, err := lb.GetById(ctx, []string{"99", "91", "90", "89", "80"})
		if err != nil {
			fmt.Printf("failed to rand: %v\n", err)
			return
		}
		fmt.Println(r)
	}

	{
		r, err := lb.GetById(ctx, []string{"80"})
		if err != nil {
			fmt.Printf("failed to rand: %v\n", err)
			return
		}
		fmt.Println(r)
	}
}

func TestLeaderboardGetById(t *testing.T) {
	ctx := context.Background()

	lb := getTestLeaderboard(
		ctx,
		WithExpire(time.Minute),
		WithCapacity(10),
		//WithNoInfo(),
		//WithNotTrim(),
	)

	{
		r, err := lb.SetOne(ctx, Entry{Id: "1", Score: 1}, WithSetOnlyAdd(true))
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}
	{
		r, err := lb.SetOne(ctx, Entry{Id: "1", Score: 2}, WithSetOnlyAdd(true))
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}
	{
		r, err := lb.SetOne(ctx, Entry{Id: "1", Score: 2}, WithSetOnlyUpdate(true))
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}
	{
		r, err := lb.SetOne(ctx, Entry{Id: "1", Score: 3})
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}
	{
		r, err := lb.SetOne(ctx, Entry{Id: "1", Score: 3}, WithSetIncrBy(true))
		if err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
		fmt.Println(r)
	}

	{
		r, err := lb.GetById(ctx, []string{"1"})
		if err != nil {
			fmt.Printf("failed to get: %v\n", err)
			return
		}
		fmt.Println(r)
	}

	{
		r, err := lb.GetById(ctx, []string{"2"})
		if err != nil {
			fmt.Printf("failed to get: %v\n", err)
			return
		}
		fmt.Println(r)
	}
}

/*
func TestLeaderboardRandomByScore(t *testing.T) {
	ctx := context.Background()

	lb := getTestLeaderboard(ctx)

	for i := 1; i <= 10; i++ {
		n := rand.Int63n(100)
		s := fmt.Sprintf("%d", n)
		if err := lb.Add(ctx, &Entry{Id: s, Info: s, Score: n}); err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
	}

	r, err := lb.RandByScore(ctx, RandByScoreArg{Min: 10, Max: 50, Count: 3})
	if err != nil {
		fmt.Printf("failed to rand: %v\n", err)
		return
	}

	fmt.Println(r)
}
*/
