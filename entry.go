package rank

import (
	"encoding/json"

	"github.com/ntons/tongo/tunsafe"
)

type Entry struct {
	Rank  int    `json:"rank" msgpack:"rank"`
	Id    string `json:"id" msgpack:"id"`
	Info  string `json:"info" msgpack:"info"`
	Score int64  `json:"score" msgpack:"score"`
}

type E = Entry

func (e Entry) String() string {
	b, _ := json.Marshal(&e)
	return tunsafe.BytesToString(b)
}
