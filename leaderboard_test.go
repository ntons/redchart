package ranking

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestLeaderboard(t *testing.T) {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	err := LoadScripts(ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	lb := GetLeaderboard(client, "leaderboard", &Options{
		Capacity: 5,
		ExpireAt: time.Now().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = lb.SetScore(ctx,
		&E{Id: "id1", Score: 1000},
		&E{Id: "id2", Score: 2000},
		&E{Id: "id3", Score: 3000},
		&E{Id: "id4", Score: 4000},
		&E{Id: "id5", Score: 5000},
		&E{Id: "id6", Score: 6000},
		&E{Id: "id7", Score: 7000},
		&E{Id: "id8", Score: 8000},
	); err != nil {
		t.Log(luaSetScore)
		t.Fatal(err)
	}
	if err = lb.IncScore(ctx,
		&E{Id: "id1", Score: 1000},
		&E{Id: "id2", Score: 2000},
		&E{Id: "id3", Score: 3000},
		&E{Id: "id4", Score: 4000},
		&E{Id: "id5", Score: 5000},
		&E{Id: "id6", Score: 6000},
		&E{Id: "id7", Score: 7000},
		&E{Id: "id8", Score: 8000},
	); err != nil {
		t.Log(luaIncScore)
		t.Fatal(err)
	}

	r, err := lb.GetRange(ctx, 0, 10)
	if err != nil {
		t.Log(luaGetRange)
		t.Fatal(err)
	}
	fmt.Println(r)

	if e, err := lb.GetById(ctx, "id1"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(e)
	}
	if e, err := lb.GetById(ctx, "id6"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(e)
	}
}
