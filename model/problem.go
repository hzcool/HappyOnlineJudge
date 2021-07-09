package model

import (
	"time"
)

type Problem struct {
	ID            int64                  `json:"id" xorm:"pk autoincr"`
	CreatedAt     time.Time              `json:"created_at" xorm:"created"`
	Author        string                 `json:"author" xorm:"index"`
	Source        string                 `json:"source" xorm:"varchar(32)"`
	Title         string                 `json:"title" xorm:"varchar(64)"`
	Background    string                 `json:"background" xorm:"text"`
	Statement     string                 `json:"statement" xorm:"text"`
	Input         string                 `json:"input" xorm:"text"`
	Output        string                 `json:"output" xorm:"text"`
	ExamplesIn    string                 `json:"examples_in" xorm:"text"`  //json
	ExamplesOut   string                 `json:"examples_out" xorm:"text"` //json
	Hint          string                 `json:"hint" xorm:"text"`
	TimeLimit     uint                   `json:"time_limit"`
	MemoryLimit   uint                   `json:"memory_limit"`
	IsOpen        bool                   `json:"is_open" xorm:"index"`
	Index         string                 `json:"index" xorm:"index unique"`
	IsSpj         bool                   `json:"is_spj" xorm:"default 0"`
	SpjType       string                 `json:"spj_type" xorm:"varchar(10)"`
	AcceptedCount uint                   `json:"accepted_count" xorm:"default 0"`
	AllCount      uint                   `json:"all_count" xorm:"default 0"`
	Tags          string                 `json:"tags" xorm:"default '[]'"`
	Testdatas     map[string]interface{} `json:"testdatas"`
}
