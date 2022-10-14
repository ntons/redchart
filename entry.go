package redchart

import "fmt"

// Entry代表排行榜中的一条数据
// Entry.Rank 从0开始的榜单排名
// Entry.Id 对象ID
// Entry.Score 对象分数，排名依据
// Entry.Info 对象的详细信息
type Entry struct {
	Rank  int32  `json:"rank" msgpack:"rank"`
	Id    string `json:"id" msgpack:"id"`
	Info  string `json:"info" msgpack:"info"`
	Score int64  `json:"score" msgpack:"score"`
}

func (e Entry) String() string {
	return fmt.Sprintf("%v,%v,%v,%v", e.Rank, e.Id, e.Score, e.Info)
}

type E = Entry
