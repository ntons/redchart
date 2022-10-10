package redchart

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestLeaderboardRandomByScore(t *testing.T) {
	ctx := context.Background()

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 9})

	cli.Del(ctx, "*")

	lb := GetLeaderboard(cli, "leaderboardtest", WithExpire(time.Minute))

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

func TestLeaderboardCapacity(t *testing.T) {
	ctx := context.Background()

	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 9})

	cli.Del(ctx, "*")

	lb := GetLeaderboard(
		cli,
		"leaderboardtest",
		WithExpire(time.Minute),
		WithCapacity(10),
		WithNoInfo(),
		//WithNotTrim(),
	)

	for i := int64(1); i <= 100; i++ {
		s := fmt.Sprintf("%d", i)
		if err := lb.Add(ctx, &Entry{Id: s, Info: s, Score: i}); err != nil {
			fmt.Printf("failed to add: %v\n", err)
			return
		}
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
		r, err := lb.GetById(ctx, "99", "91", "80")
		if err != nil {
			fmt.Printf("failed to rand: %v\n", err)
			return
		}
		fmt.Println(r)
	}
}
