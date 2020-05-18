package rank

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestVector(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	err := ScriptLoad(client)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	vec := GetVector(client, "vector", &Options{
		Capacity:   5,
		ExpireAt:   time.Now().Add(time.Minute),
		IdleExpire: 10 * time.Second,
	})
	if err = vec.Append(ctx,
		&E{Id: "id1"},
		&E{Id: "id2"},
		&E{Id: "id3"},
		&E{Id: "id4"},
		&E{Id: "id5"},
		&E{Id: "id6"},
		&E{Id: "id7"},
		&E{Id: "id8"},
	); err != nil {
		t.Log(luaSetScore)
		t.Fatal(err)
	}

	if r, err := vec.GetRange(ctx, 0, 10); err != nil {
		t.Log(luaGetRange)
		t.Fatal(err)
	} else {
		fmt.Println(r)
	}

	if err := vec.SwapById(ctx, "id1", "id3"); err != nil {
		t.Fatal(err)
	}

	if r, err := vec.GetRange(ctx, 0, 10); err != nil {
		t.Log(luaGetRange)
		t.Fatal(err)
	} else {
		fmt.Println(r)
	}

	if e, err := vec.GetById(ctx, "id1"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(e)
	}
	if e, err := vec.GetById(ctx, "id6"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(e)
	}
}
